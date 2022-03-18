package interp

import (
	"errors"
	"fmt"
	"math/big"
	"math/bits"
	"strconv"
	"strings"

	"tinygo.org/x/go-llvm"
)

func parseAsIntrinsic(c *constParser, call callInst) (instruction, error) {
	p, ok := call.called.val.(*offPtr)
	if !ok {
		return nil, errRuntime
	}
	obj := p.obj()
	if !obj.isFunc || call.called.raw != 0 {
		return nil, errRuntime
	}

	switch obj.name {
	case "runtime.alloc":
		return parseHeapAlloc(c, call)
	case "runtime.trackPointer":
		return nil, nil
	case "runtime.runtimePanic":
		return &runtimePanicInst{call.args[0], call.args[1], call.dbg}, nil
	case "memset":
		return &memSetInst{call.args[0], cast(i8, call.args[1]), call.args[2], call}, nil
	case "memmove":
		return &memMoveInst{call.args[0], call.args[1], call.args[2], true, call}, nil
	case "memcpy":
		return &memMoveInst{call.args[0], call.args[1], call.args[2], false, call}, nil
	default:
		switch {
		case strings.HasPrefix(obj.name, "llvm.memset."):
			if call.args[3] == smallIntValue(i1, 0) {
				return &memSetInst{call.args[0], call.args[1], call.args[2], call}, nil
			}
		case strings.HasPrefix(obj.name, "llvm.memmove."):
			if call.args[3] == smallIntValue(i1, 0) {
				return &memMoveInst{call.args[0], call.args[1], call.args[2], true, call}, nil
			}
		case strings.HasPrefix(obj.name, "llvm.memcpy."):
			if call.args[3] == smallIntValue(i1, 0) {
				return &memMoveInst{call.args[0], call.args[1], call.args[2], false, call}, nil
			}
		case strings.HasPrefix(obj.name, "llvm.lifetime.start.") || strings.HasPrefix(obj.name, "llvm.lifetime.end.") ||
			strings.HasPrefix(obj.name, "llvm.dbg."):
			return nil, nil
		}
		return nil, errRuntime
	}
}

func parseHeapAlloc(c *constParser, call callInst) (instruction, error) {
	size, layout := call.args[0], call.args[1]
	elemTy, err := c.parseLayout(layout)
	if err != nil {
		return nil, err
	}
	return &heapAllocInst{size, elemTy, call}, nil
}

func (p *constParser) parseLayout(layout value) (typ, error) {
	if t, ok := p.layouts[layout]; ok {
		return t, nil
	}
	var t typ
	var err error
	switch val := layout.val.(type) {
	case castVal:
		return p.parseLayout(value{val.val, layout.raw})

	case smallInt:
		if layout.raw == 0 {
			// Nil pointer, which means the layout is unknown.
			return nil, errRuntime
		}
		if layout.raw%2 != 1 {
			// Sanity check: the least significant bit must be set. This is how
			// the runtime can separate pointers from integers.
			return nil, errors.New("unexpected layout")
		}

		// Determine format of bitfields in the integer.
		var sizeFieldBits uint64
		switch p.uintptr {
		case i16:
			sizeFieldBits = 4
		case i32:
			sizeFieldBits = 5
		case i64:
			sizeFieldBits = 6
		default:
			panic("unknown pointer type")
		}

		// Extract fields.
		objectSizeWords := (layout.raw >> 1) & (1<<sizeFieldBits - 1)
		bitmap := new(big.Int).SetUint64(layout.raw >> (1 + sizeFieldBits))

		// Parse the layout.
		t, err = p.layoutFromBits(objectSizeWords, bitmap)

	case *offPtr:
		obj := val.obj()
		if !obj.isConst || layout.raw != 0 {
			// This isn't a direct reference to a constant.
			return nil, errRuntime
		}

		// Load the layout data from the object.
		err = obj.parseInit(p)
		if err != nil {
			return nil, err
		}
		var size uint64
		if sizeV, err := obj.init.load(p.uintptr, 0); err != nil {
			return nil, err
		} else if _, ok := sizeV.val.(smallInt); !ok {
			return nil, errRuntime
		} else {
			size = sizeV.raw
		}
		rawBytes := make([]byte, (size+7)/8)
		off := p.uintptr.bytes()
		for i := range rawBytes {
			v, err := obj.init.load(i8, off+uint64(i))
			if err != nil {
				return nil, err
			}
			if _, ok := v.val.(smallInt); !ok {
				return nil, errRuntime
			}
			rawBytes[i] = byte(v.raw)
		}
		bitmap := new(big.Int).SetBytes(rawBytes)

		// Parse the layout.
		t, err = p.layoutFromBits(size, bitmap)

	default:
		return nil, err
	}
	if err != nil {
		return nil, err
	}

	p.layouts[layout] = t
	return t, nil
}

func (p *constParser) layoutFromBits(size uint64, bits *big.Int) (typ, error) {
	if bits.BitLen() == 0 {
		// There are no pointers here.
		return i8, nil
	}

	// Select element types based on the bitmap.
	uintptr := p.uintptr
	usize := uintptr.bytes()
	partSize := uint64(p.ptrAlign)
	ptrSplit := usize / partSize
	rawT, ptrT := p.ctx.IntType(int(8*partSize)), llvm.PointerType(p.ctx.Int8Type(), 0)
	buf := make([]llvm.Type, size)[:0]
	for i := uint64(0); i < size; {
		if bits.Bit(int(i)) != 0 {
			buf = append(buf, ptrT)
			i += ptrSplit
		} else {
			buf = append(buf, rawT)
			i++
		}
	}

	// Compact the types by packing adjacent elements into arrays.
	// This is not necessary, but it makes the type definitions for large structures much easier to read.
	old := buf
	buf = buf[:0]
	for i := 0; i < len(old); {
		start := i
		ty := old[i]
		for i < len(old) && old[i] == ty {
			i++
		}
		n := i - start
		if n == 1 {
			buf = append(buf, ty)
		} else {
			buf = append(buf, llvm.ArrayType(ty, n))
		}
	}

	// Merge the element types.
	if len(buf) == 1 {
		return p.typ(buf[0])
	}
	return p.typ(p.ctx.StructType(buf, false))
}

// heapAllocInst is an intrinsic instruction that does a heap allocation.
type heapAllocInst struct {
	// size is the size of the allocation in bytes.
	size value

	// elemTy is the element type of the resulting allocation.
	elemTy typ

	// callInst is a raw call to runtime.alloc.
	callInst
}

func (i *heapAllocInst) exec(state *execState) error {
	size := i.size.resolve(state.locals())
	if _, ok := size.val.(smallInt); !ok {
		// The size is not constant.
		// Do the allocation at runtime.
		return i.callInst.exec(state)
	}
	elemSize := i.elemTy.bytes()
	if (elemSize != 0 && size.raw%elemSize != 0) || (elemSize == 0 && size.raw != 0) {
		return fmt.Errorf("requested size of %d bytes is not a multiple of element %s size %d", size.raw, i.elemTy.String(), elemSize)
	}
	if size.raw >= 1<<32 {
		return errRevert{fmt.Errorf("allocation of %d bytes is too big", size.raw)}
	}
	var n uint32
	if size.raw != 0 {
		n = uint32(size.raw / elemSize)
	}
	var alignScale uint
	if n != 0 && state.cp.ptrAlign != 1 {
		alignScale = uint(bits.Len64(size.raw) - 1)
		if alignScale > 4 {
			// TODO: find a better way to select max alignment.
			// This is the maximum alignment which any target seems to ever expect.
			alignScale = 4
		}
	}
	var ty typ
	if n == 1 {
		ty = i.elemTy
	} else {
		ty = array(i.elemTy, n)
	}
	mem := makeZeroMem(size.raw)
	id := state.nextObjID
	state.nextObjID++
	obj := &memObj{
		ptrTy:      i.sig.ret.(ptrType),
		name:       "interp.alloc." + strconv.FormatUint(id, 10),
		id:         id,
		unique:     size.raw != 0,
		size:       size.raw,
		alignScale: alignScale,
		ty:         ty,
		dbg:        i.dbg,
		version:    state.version,
		init:       mem,
		data:       mem,
	}
	state.stack = append(state.stack, obj.ptr(0))
	return nil
}

type runtimePanicInst struct {
	base, len value
	dbg       llvm.Metadata
}

var _ instruction = (*runtimePanicInst)(nil)

func (i *runtimePanicInst) result() typ {
	return nil
}

func (i *runtimePanicInst) exec(state *execState) error {
	base := i.base.resolve(state.locals())
	len := i.len.resolve(state.locals())
	str, err := state.loadDbgString(base, len)
	if err != nil {
		println("failed to load panic str", err.Error())
		return fmt.Errorf("runtime panic (%s, %s)%s", base.String(), len.String(), dbgSuffix(i.dbg))
	}
	return fmt.Errorf("runtime panic %q%s", str, dbgSuffix(i.dbg))
}

func (i *runtimePanicInst) runtime(*rtGen) error {
	return errors.New("cannot runtime panic at runtime")
}

func (i *runtimePanicInst) String() string {
	return fmt.Sprintf("runtime panic %s, %s", i.base.String(), i.len.String())
}

func (s *execState) loadDbgString(base, len value) (string, error) {
	p, ok := base.val.(*offPtr)
	if !ok {
		return "", errRuntime
	}
	_, ok = len.val.(smallInt)
	if !ok {
		return "", errRuntime
	}
	obj := p.obj()
	if obj.isExtern || obj.isFunc {
		return "", errRuntime
	}
	err := obj.parseInit(&s.cp)
	if err != nil {
		return "", err
	}
	if len.raw > 100*1024 || len.raw <= -base.raw {
		// This does not seem like a great idea.
		return "", errRuntime
	}
	data := make([]byte, len.raw)
	for i := uint64(0); i < len.raw; i++ {
		v, err := obj.data.load(i8, base.raw+i)
		if err != nil {
			return "", err
		}
		_, ok := v.val.(smallInt)
		if !ok {
			return "", errRuntime
		}
		data[i] = byte(v.raw)
	}
	return string(data), nil
}

// memSetInst is an instruction to fill a byte range with a specified value.
type memSetInst struct {
	// dst is a pointer to the start of the range to be filled.
	dst value

	// val is the raw byte value to fill the range with.
	val value

	// len is the length of the range to fill.
	len value

	// callInst is a raw call to a memory set intrinsic or function.
	callInst
}

func (i *memSetInst) exec(state *execState) error {
	err := i.tryExec(state)
	if err != nil {
		if isRuntimeOrRevert(err) {
			return i.callInst.exec(state)
		}
		return err
	}
	return nil
}

func (i *memSetInst) tryExec(state *execState) error {
	locals := state.locals()

	len := i.len.resolve(locals)
	_, ok := len.val.(smallInt)
	if !ok {
		return errRuntime
	}
	dst := i.dst.resolve(locals)
	if len.raw != 0 {
		ptr, ok := dst.val.(*offPtr)
		if !ok {
			return errRuntime
		}
		obj := ptr.obj()
		if len.raw > obj.size || dst.raw > obj.size-len.raw {
			return errRuntime
		}
		if obj.isConst {
			return errUB
		}
		v := i.val.resolve(locals)
		_, ok = v.val.(smallInt)
		if !ok {
			return errRuntime
		}
		err := obj.parseInit(&state.cp)
		if err != nil {
			return err
		}
		node, err := obj.data.memSet(dst.raw, len.raw, byte(v.raw), state.version)
		if err != nil {
			return err
		}
		if obj.version < state.version {
			state.oldMem = append(state.oldMem, memSave{
				obj:     obj,
				tree:    obj.data,
				version: obj.version,
			})
		}
		obj.data = node
		obj.version = state.version
	}
	if i.result() != nil {
		// The C memset function returns the destination argument.
		// The LLVM memset intrinsics do not.
		state.stack = append(state.stack, dst)
	}
	return nil
}

// memMoveInst is an instruction to move a range of data from one location to another.
type memMoveInst struct {
	dst, src value
	len      value

	allowOverlap bool

	callInst
}

func (i *memMoveInst) exec(state *execState) error {
	err := i.tryExec(state)
	if err != nil {
		if isRuntimeOrRevert(err) {
			return i.callInst.exec(state)
		}
		return err
	}
	return nil
}

func (i *memMoveInst) tryExec(state *execState) error {
	locals := state.locals()
	len := i.len.resolve(locals)
	if _, ok := len.val.(smallInt); !ok {
		return errRuntime
	}
	if len.raw == 0 {
		return nil
	}

	dst := i.dst.resolve(locals)
	dstPtr, ok := dst.val.(*offPtr)
	if !ok {
		return errRuntime
	}
	src := i.src.resolve(locals)
	srcPtr, ok := src.val.(*offPtr)
	if !ok {
		return errRuntime
	}

	dstObj := dstPtr.obj()
	if err := dstObj.parseInit(&state.cp); err != nil {
		return err
	}
	srcObj := srcPtr.obj()
	if err := srcObj.parseInit(&state.cp); err != nil {
		return err
	}
	if len.raw > srcObj.size || len.raw > dstObj.size || src.raw > srcObj.size-len.raw || dst.raw > dstObj.size-len.raw {
		return errRuntime
	}
	if dstObj.isConst {
		return errUB
	}

	// Verify that the source and destination do not overlap.
	if !i.allowOverlap && srcObj == dstObj && (src.raw == dst.raw || (src.raw < dst.raw && dst.raw-src.raw < len.raw) || (dst.raw < src.raw && src.raw-dst.raw < len.raw)) {
		return errUB
	}

	// Create a new version to force a copy of the destination.
	version := state.nextVersion
	state.nextVersion++

	node, err := srcObj.data.copyTo(dstObj.data, dst.raw, src.raw, len.raw, version)
	if err != nil {
		return err
	}
	if dstObj.version < state.version {
		state.oldMem = append(state.oldMem, memSave{
			obj:     dstObj,
			tree:    dstObj.data,
			version: dstObj.version,
		})
	}
	dstObj.data = node
	dstObj.version = state.version

	if i.result() != nil {
		// The C memcpy/memmove functions return the destination argument.
		// The LLVM memcpy/memmove intrinsics do not.
		state.stack = append(state.stack, dst)
	}

	return nil
}

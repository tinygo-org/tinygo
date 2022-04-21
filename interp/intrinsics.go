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
		return &heapAllocInst{call.args[0], call.args[1], call}, nil
	case "runtime.trackPointer":
		return nil, nil
	case "runtime.runtimePanic":
		return &runtimePanicInst{call.args[0], call.args[1], call.dbg}, nil
	case "runtime.typeAssert":
		return &typeAssertInst{call.args[0], cast(call.args[0].typ().(nonAggTyp), call.args[1]), call}, nil
	case "(reflect.rawType).elem":
		return &typeElemInst{cast(c.uintptr, call.args[0]), call}, nil
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
		case strings.HasSuffix(obj.name, "$invoke"):
			return &interfaceInvokeInst{cast(pointer(defaultAddrSpace, c.uintptr), call.args[len(call.args)-2]), call}, nil
		case strings.HasPrefix(obj.name, "llvm.lifetime.start.") || strings.HasPrefix(obj.name, "llvm.lifetime.end.") ||
			strings.HasPrefix(obj.name, "llvm.dbg."):
			return nil, nil
		}
		return nil, errRuntime
	}
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
		return nil, errRuntime
	}
	if err != nil {
		return nil, err
	}

	p.layouts[layout] = t
	return t, nil
}

func (p *constParser) layoutFromBits(size uint64, bitmap *big.Int) (typ, error) {
	if bitmap.BitLen() == 0 {
		// There are no pointers here.
		return i8, nil
	}

	// Select element types based on the bitmap.
	uintptr := p.uintptr
	usize := uintptr.bytes()
	partSize := uint64(1) << p.align(p.uintptr)
	ptrSplit := usize / partSize
	ptrT := llvm.PointerType(p.ctx.Int8Type(), 0)
	rawTypes := make([]llvm.Type, 4)[:0]
	maxTy := i64
	if (size*partSize)%8 != 0 {
		// The maximum element alignment is capped by the alignment of the size.
		maxTy = iType(8*partSize) << bits.TrailingZeros64(size)
	}
	for i := iType(8 * partSize); i <= maxTy; i *= 2 {
		rawTypes = append(rawTypes, p.ctx.IntType(int(i)))
	}
	buf := make([]llvm.Type, size)[:0]
	for i := uint64(0); i < size; {
		if bitmap.Bit(int(i)) != 0 {
			buf = append(buf, ptrT)
			i += ptrSplit
		} else {
			scale := 0
			for scale < len(rawTypes)-1 && i+(1<<scale) < size && i&((1<<(scale+1))-1) == 0 {
				ok := true
				for j := i + (1 << scale); j > i+((1<<scale)>>1); j-- {
					if bitmap.Bit(int(j)) != 0 {
						ok = false
						break
					}
				}
				if !ok {
					break
				}
				scale++
			}
			buf = append(buf, rawTypes[scale])
			i += 1 << scale
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
	t, err := p.typ(p.ctx.StructType(buf, false))
	if err != nil {
		return nil, err
	}
	if assert {
		gotSize := t.bytes()
		if gotSize != partSize*size {
			panic("size mismatch")
		}
	}
	return t, nil
}

// heapAllocInst is an intrinsic instruction that does a heap allocation.
type heapAllocInst struct {
	// size is the size of the allocation in bytes.
	size value

	// layout is the layout pointer of the allocation.
	layout value

	// callInst is a raw call to runtime.alloc.
	callInst
}

func (i *heapAllocInst) exec(state *execState) error {
	locals := state.locals()
	size := i.size.resolve(locals)
	if _, ok := size.val.(smallInt); !ok {
		// The size is not constant.
		// Do the allocation at runtime.
		return i.callInst.exec(state)
	}
	layout := i.layout.resolve(locals)
	elemTy, err := state.cp.parseLayout(layout)
	switch err {
	case nil:
	case errRuntime:
		// The layout is not known.
		// Do the allocation at runtime.
		return i.callInst.exec(state)
	default:
		panic(err)
		//return err
	}
	if size.raw >= 1<<32 {
		return errRevert{fmt.Errorf("allocation of %d bytes is too big", size.raw)}
	}
	if elemTy == i8 && size.raw > 1 && state.cp.align(state.cp.uintptr) != 0 {
		// Use bigger integers to get maximum alignment.
		scale := bits.TrailingZeros64(size.raw)
		if scale > 3 {
			scale = 3
		}
		elemTy = i8 << scale
	}
	elemSize := elemTy.bytes()
	if (elemSize != 0 && size.raw%elemSize != 0) || (elemSize == 0 && size.raw != 0) {
		return fmt.Errorf("requested size of %d bytes is not a multiple of element %s size %d", size.raw, elemTy.String(), elemSize)
	}
	var n uint32
	if size.raw != 0 {
		n = uint32(size.raw / elemSize)
	}
	var ty typ
	if n == 1 {
		ty = elemTy
	} else {
		ty = array(elemTy, n)
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
		alignScale: state.cp.align(elemTy),
		ty:         ty,
		dbg:        i.dbg,
		version:    state.version,
		init:       mem,
		data:       mem,
	}
	state.stack = append(state.stack, obj.ptr(0))
	return nil
}

func (i *heapAllocInst) String() string {
	return "intrinsic " + i.callInst.String()
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

func (i *memSetInst) String() string {
	return "intrinsic " + i.callInst.String()
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

func (i *memMoveInst) String() string {
	return "intrinsic " + i.callInst.String()
}

// typeAssertInst checks if an actual type matches an asserted type exactly.
// This runtime.typeAssert call would normally be implemented by the interface lowering pass.
type typeAssertInst struct {
	actualType, assertedType value

	callInst
}

func (i *typeAssertInst) exec(state *execState) error {
	err := i.tryExec(state)
	if err != nil {
		if isRuntimeOrRevert(err) {
			return i.callInst.exec(state)
		}
		return err
	}
	return nil
}

func (i *typeAssertInst) tryExec(state *execState) error {
	locals := state.locals()
	actualType := i.actualType.resolve(locals)
	assertedType := i.assertedType.resolve(locals)

	type kind uint8
	const (
		idk kind = iota
		null
		ty
	)

	var actual, asserted string
	var actualKind, assertedKind kind
	switch val := actualType.val.(type) {
	case smallInt:
		if actualType.raw == 0 {
			actualKind = null
		}
	case *offAddr:
		obj := val.obj()
		if strings.HasPrefix(obj.name, "reflect/types.type:") {
			actual, actualKind = strings.TrimPrefix(obj.name, "reflect/types.type:"), ty
		}
	}
	switch val := assertedType.val.(type) {
	case smallInt:
		if assertedType.raw == 0 {
			assertedKind = null
		}
	case *offAddr:
		obj := val.obj()
		if strings.HasPrefix(obj.name, "reflect/types.typeid:") {
			asserted, assertedKind = strings.TrimPrefix(obj.name, "reflect/types.typeid:"), ty
		}
	}

	type kinds struct{ actual, asserted kind }
	switch (kinds{actualKind, assertedKind}) {
	case kinds{null, null}:
		state.stack = append(state.stack, boolValue(true))
		return nil
	case kinds{ty, ty}:
		state.stack = append(state.stack, boolValue(actual == asserted))
		return nil
	}

	return errRuntime
}

func (i *typeAssertInst) String() string {
	return "intrinsic " + i.callInst.String()
}

// typeElemInst looks up the element of the provided type.
type typeElemInst struct {
	typeCodeAddr value

	callInst
}

func (i *typeElemInst) exec(state *execState) error {
	err := i.tryExec(state)
	if err != nil {
		if isRuntimeOrRevert(err) {
			return i.callInst.exec(state)
		}
		return err
	}
	return nil
}

func (i *typeElemInst) tryExec(state *execState) error {
	typeCodeAddr := i.typeCodeAddr.resolve(state.locals())
	ptr, ok := typeCodeAddr.val.(*offAddr)
	if !ok || typeCodeAddr.raw != 0 {
		if debug {
			println("cannot use type code addr", typeCodeAddr.String())
		}
		return errRuntime
	}
	obj := ptr.obj()
	const prefix = "reflect/types.type:"
	if !strings.HasPrefix(obj.name, prefix) {
		if debug {
			println("type elem code missing prefix", obj.name)
		}
		return errRuntime
	}
	id := strings.TrimPrefix(obj.name, prefix)
	class := id[:strings.IndexByte(id, ':')]
	value := id[len(class)+1:]
	if class == "named" {
		// Get the underlying type.
		class = value[:strings.IndexByte(value, ':')]
		value = value[len(class)+1:]
	}
	// Elem() is only valid for certain type classes.
	switch class {
	case "chan", "pointer", "slice", "array":
	default:
		return fmt.Errorf("(reflect.Type).Elem() called on %s type", class)
	}
	err := obj.parseInit(&state.cp)
	if err != nil {
		if debug {
			println("type elem parse init failed", err.Error())
		}
		return err
	}
	v, err := obj.data.load(typeCodeAddr.typ(), 0)
	if err != nil {
		if debug {
			println("type elem load failed", err.Error())
		}
		return err
	}
	state.stack = append(state.stack, v)
	return nil
}

func (i *typeElemInst) String() string {
	return "intrinsic " + i.callInst.String()
}

type interfaceInvokeInst struct {
	typeCodePtr value

	callInst
}

func (i *interfaceInvokeInst) exec(state *execState) error {
	typeCodePtr := i.typeCodePtr.resolve(state.locals())
	ptr, ok := typeCodePtr.val.(*offPtr)
	if !ok || typeCodePtr.raw != 0 {
		return i.callInst.exec(state)
	}
	obj := ptr.obj()
	if !strings.HasPrefix(obj.name, "reflect/types.type:") {
		return errors.New("bad type code name")
	}
	err := obj.parseInit(&state.cp)
	if err != nil {
		return err
	}
	methodSet, err := obj.data.load(obj.ptrTy, obj.ty.(*structType).fields[2].offset)
	if err != nil {
		return err
	}
	thunkSig, err := i.called.val.(*offPtr).obj().parseSig(&state.cp)
	if err != nil {
		return err
	}
	sig := state.cp.globalsByName[thunkSig.invokeName]
	if sig == nil {
		return errors.New("missing invoke signature")
	}
	methodSetObj := methodSet.val.(*offPtr).obj()
	err = methodSetObj.parseInit(&state.cp)
	if err != nil {
		return err
	}
	marr := methodSetObj.ty.(arrType)
	interfaceMethodInfo := marr.of.(*structType)
	pairWidth := interfaceMethodInfo.bytes()
	for j := uint32(0); j < marr.n; j++ {
		mSig, err := methodSetObj.data.load(interfaceMethodInfo.fields[0].ty, uint64(j)*pairWidth+interfaceMethodInfo.fields[0].offset)
		if err != nil {
			return err
		}
		if mSig == sig.ptr(0) {
			fnPtr, err := methodSetObj.data.load(interfaceMethodInfo.fields[1].ty, uint64(j)*pairWidth+interfaceMethodInfo.fields[1].offset)
			if err != nil {
				return err
			}
			ptr, ok := fnPtr.val.(*offAddr)
			if !ok || !ptr.isFunc {
				return fmt.Errorf("cannot call %s", fnPtr.String())
			}
			obj := ptr.obj()
			sig, err := obj.parseSig(&state.cp)
			if err != nil {
				return err
			}
			args := make([]value, len(i.args)-1)
			copy(args, i.args[:len(i.args)-2])
			args[len(args)-1] = undefValue(pointer(defaultAddrSpace, state.cp.uintptr))
			return (&callInst{cast(i.callInst.called.typ().(nonAggTyp), fnPtr), args, sig, i.dbg, i.recursiveRevert}).exec(state)
		}
	}
	return errors.New("no matching method")
}

func (i *interfaceInvokeInst) String() string {
	return "intrinsic " + i.callInst.String()
}

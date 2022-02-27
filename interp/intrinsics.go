package interp

import (
	"errors"
	"fmt"
	"math/big"
	"math/bits"
	"strconv"

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
	default:
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
			return nil, todo("unknown allocation layout")
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
		return fmt.Errorf("allocation of %d bytes is too big", size.raw)
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

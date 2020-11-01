package gctype

import (
	"errors"
	"fmt"
	"math/big"

	"tinygo.org/x/go-llvm"
)

// getPointerBitmap scans the given LLVM type for pointers and sets bits in a
// bigint at the word offset that contains a pointer. This scan is recursive.
func getPointerBitmap(targetData llvm.TargetData, typ llvm.Type, name string) *big.Int {
	alignment := targetData.PrefTypeAlignment(llvm.PointerType(typ.Context().Int8Type(), 0))
	switch typ.TypeKind() {
	case llvm.IntegerTypeKind, llvm.FloatTypeKind, llvm.DoubleTypeKind:
		return big.NewInt(0)
	case llvm.PointerTypeKind:
		return big.NewInt(1)
	case llvm.StructTypeKind:
		ptrs := big.NewInt(0)
		for i, subtyp := range typ.StructElementTypes() {
			subptrs := getPointerBitmap(targetData, subtyp, name)
			if subptrs.BitLen() == 0 {
				continue
			}
			offset := targetData.ElementOffset(typ, i)
			if offset%uint64(alignment) != 0 {
				panic("precise GC: global contains unaligned pointer: " + name)
			}
			subptrs.Lsh(subptrs, uint(offset)/uint(alignment))
			ptrs.Or(ptrs, subptrs)
		}
		return ptrs
	case llvm.ArrayTypeKind:
		subtyp := typ.ElementType()
		subptrs := getPointerBitmap(targetData, subtyp, name)
		ptrs := big.NewInt(0)
		if subptrs.BitLen() == 0 {
			return ptrs
		}
		elementSize := targetData.TypeAllocSize(subtyp)
		for i := 0; i < typ.ArrayLength(); i++ {
			ptrs.Lsh(ptrs, uint(elementSize)/uint(alignment))
			ptrs.Or(ptrs, subptrs)
		}
		return ptrs
	default:
		panic("unknown type kind: " + name)
	}
}

// NewTyper creates a Typer.
func NewTyper(ctx llvm.Context, mod llvm.Module, td llvm.TargetData) *Typer {
	ptr := llvm.PointerType(ctx.Int8Type(), 0)
	return &Typer{
		tcache:   make(map[llvm.Type]llvm.Value),
		td:       td,
		ctx:      ctx,
		mod:      mod,
		uintptr:  ctx.IntType(int(td.TypeSizeInBits(ptr))),
		i8:       ctx.Int8Type(),
		ptrSize:  td.TypeAllocSize(ptr),
		ptrAlign: uint64(td.ABITypeAlignment(ptr)),
	}
}

// Typer creates GC types.
type Typer struct {
	// tcache is a cache of GC types by LLVM type.
	tcache map[llvm.Type]llvm.Value

	// td is the target platform data.
	td llvm.TargetData

	ctx llvm.Context

	mod llvm.Module

	uintptr, i8 llvm.Type

	ptrSize, ptrAlign uint64
}

// Create a GC type for the given LLVM type.
func (t *Typer) Create(typ llvm.Type) (llvm.Value, error) {
	// Check the cache before attempting to create the global.
	if g, ok := t.tcache[typ]; ok {
		return g, nil
	}

	// Find the type size.
	size := t.td.TypeAllocSize(typ)

	// Compute a pointer bitmap.
	// TODO: clean this up and maybe use error handling?
	b := getPointerBitmap(t.td, typ, "")
	if b.Cmp(big.NewInt(0)) == 0 {
		// The type has no pointers.
		return llvm.ConstNull(llvm.PointerType(t.uintptr, 0)), nil
	}

	// Use some limited sanity-checking.
	align := uint64(t.td.ABITypeAlignment(typ))
	switch {
	case size < t.ptrSize:
		return llvm.Value{}, errors.New("type has pointers but is smaller than a pointer")
	case align%t.ptrAlign != 0:
		return llvm.Value{}, errors.New("alignment of pointery type is not a multiple of pointer alignment")
	case size%align != 0:
		return llvm.Value{}, errors.New("type violates the array alignment invariant")
	}

	// Convert size into increments of pointer-align.
	size /= t.ptrAlign

	// Create a global for the type.
	g := t.createGlobal(size, b.Bytes())

	// Save the global to the cache.
	t.tcache[typ] = g

	return g, nil
}

func (t *Typer) createGlobal(size uint64, layout []byte) llvm.Value {
	// TODO: compression?

	// Generate the name of the global.
	name := fmt.Sprintf("tinygo.gc.type.%d.%x", size, layout)

	// Create the global if it does not exist.
	g := t.mod.NamedGlobal(name)
	if g.IsNil() {
		// Convert the encoded layout to a byte array.
		bitmapValues := make([]llvm.Value, len(layout))
		for i, b := range layout {
			bitmapValues[len(layout)-i-1] = llvm.ConstInt(t.i8, uint64(b), false)
		}
		bitmapArray := llvm.ConstArray(t.i8, bitmapValues)

		// Construct a tuple of the size + the array.
		tuple := t.ctx.ConstStruct([]llvm.Value{
			llvm.ConstInt(t.uintptr, size, false),
			bitmapArray,
		}, false)

		// Create a global constant initialized with the tuple.
		g = llvm.AddGlobal(t.mod, tuple.Type(), name)
		g.SetInitializer(tuple)
		g.SetGlobalConstant(true)
		g.SetUnnamedAddr(true)
		g.SetLinkage(llvm.InternalLinkage)
	}

	// Get a pointer to the size component of the global.
	// This is used because different globals will end up with different sizes.
	g = llvm.ConstBitCast(g, llvm.PointerType(t.uintptr, 0))

	return g
}

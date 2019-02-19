package compiler

import (
	"math/big"

	"tinygo.org/x/go-llvm"
)

func (c *Compiler) addGlobalsBitmap() {
	if c.mod.NamedGlobal("runtime.trackedGlobalsStart").IsNil() {
		return // nothing to do: no GC in use
	}

	var trackedGlobals []llvm.Value
	var trackedGlobalTypes []llvm.Type
	for global := c.mod.FirstGlobal(); !global.IsNil(); global = llvm.NextGlobal(global) {
		if global.IsDeclaration() {
			continue
		}
		typ := global.Type().ElementType()
		ptrs := c.getPointerBitmap(typ, global.Name())
		if ptrs.BitLen() == 0 {
			continue
		}
		trackedGlobals = append(trackedGlobals, global)
		trackedGlobalTypes = append(trackedGlobalTypes, typ)
	}

	//
	globalsBundleType := c.ctx.StructType(trackedGlobalTypes, false)
	globalsBundle := llvm.AddGlobal(c.mod, globalsBundleType, "tinygo.trackedGlobals")
	globalsBundle.SetLinkage(llvm.InternalLinkage)
	globalsBundle.SetUnnamedAddr(true)
	initializer := llvm.Undef(globalsBundleType)
	for i, global := range trackedGlobals {
		initializer = llvm.ConstInsertValue(initializer, global.Initializer(), []uint32{uint32(i)})
		gep := llvm.ConstGEP(globalsBundle, []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false),
		})
		global.ReplaceAllUsesWith(gep)
		global.EraseFromParentAsGlobal()
	}
	globalsBundle.SetInitializer(initializer)

	trackedGlobalsStart := llvm.ConstPtrToInt(globalsBundle, c.uintptrType)
	c.mod.NamedGlobal("runtime.trackedGlobalsStart").SetInitializer(trackedGlobalsStart)

	alignment := c.targetData.PrefTypeAlignment(c.i8ptrType)
	trackedGlobalsLength := llvm.ConstInt(c.uintptrType, c.targetData.TypeAllocSize(globalsBundleType)/uint64(alignment), false)
	c.mod.NamedGlobal("runtime.trackedGlobalsLength").SetInitializer(trackedGlobalsLength)

	bitmapBytes := c.getPointerBitmap(globalsBundleType, "globals bundle").Bytes()
	bitmapValues := make([]llvm.Value, len(bitmapBytes))
	for i, b := range bitmapBytes {
		bitmapValues[len(bitmapBytes)-i-1] = llvm.ConstInt(c.ctx.Int8Type(), uint64(b), false)
	}
	bitmapArray := llvm.ConstArray(llvm.ArrayType(c.ctx.Int8Type(), len(bitmapBytes)), bitmapValues)
	bitmapNew := llvm.AddGlobal(c.mod, bitmapArray.Type(), "runtime.trackedGlobalsBitmap.tmp")
	bitmapOld := c.mod.NamedGlobal("runtime.trackedGlobalsBitmap")
	bitmapOld.ReplaceAllUsesWith(bitmapNew)
	bitmapNew.SetInitializer(bitmapArray)
	bitmapNew.SetName("runtime.trackedGlobalsBitmap")
}

func (c *Compiler) getPointerBitmap(typ llvm.Type, name string) *big.Int {
	alignment := c.targetData.PrefTypeAlignment(c.i8ptrType)
	switch typ.TypeKind() {
	case llvm.IntegerTypeKind, llvm.FloatTypeKind, llvm.DoubleTypeKind:
		return big.NewInt(0)
	case llvm.PointerTypeKind:
		return big.NewInt(1)
	case llvm.StructTypeKind:
		ptrs := big.NewInt(0)
		for i, subtyp := range typ.StructElementTypes() {
			subptrs := c.getPointerBitmap(subtyp, name)
			if subptrs.BitLen() == 0 {
				continue
			}
			offset := c.targetData.ElementOffset(typ, i)
			if offset%uint64(alignment) != 0 {
				panic("precise GC: global contains unaligned pointer: " + name)
			}
			subptrs.Lsh(subptrs, uint(offset)/uint(alignment))
			ptrs.Or(ptrs, subptrs)
		}
		return ptrs
	case llvm.ArrayTypeKind:
		subtyp := typ.ElementType()
		subptrs := c.getPointerBitmap(subtyp, name)
		ptrs := big.NewInt(0)
		if subptrs.BitLen() == 0 {
			return ptrs
		}
		elementSize := c.targetData.TypeAllocSize(subtyp)
		for i := 0; i < typ.ArrayLength(); i++ {
			ptrs.Lsh(ptrs, uint(elementSize)/uint(alignment))
			ptrs.Or(ptrs, subptrs)
		}
		return ptrs
	default:
		panic("unknown type kind of global: " + name)
	}
}

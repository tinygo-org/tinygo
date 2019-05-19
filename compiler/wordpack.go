package compiler

// This file contains utility functions to pack and unpack sets of values. It
// can take in a list of values and tries to store it efficiently in the pointer
// itself if possible and legal.

import (
	"tinygo.org/x/go-llvm"
)

// emitPointerPack packs the list of values into a single pointer value using
// bitcasts, or else allocates a value on the heap if it cannot be packed in the
// pointer value directly. It returns the pointer with the packed data.
func (c *Compiler) emitPointerPack(values []llvm.Value) llvm.Value {
	valueTypes := make([]llvm.Type, len(values))
	for i, value := range values {
		valueTypes[i] = value.Type()
	}
	packedType := c.ctx.StructType(valueTypes, false)

	// Allocate memory for the packed data.
	var packedAlloc, packedHeapAlloc llvm.Value
	size := c.targetData.TypeAllocSize(packedType)
	if size == 0 {
		return llvm.ConstPointerNull(c.i8ptrType)
	} else if len(values) == 1 && values[0].Type().TypeKind() == llvm.PointerTypeKind {
		return c.builder.CreateBitCast(values[0], c.i8ptrType, "pack.ptr")
	} else if size <= c.targetData.TypeAllocSize(c.i8ptrType) {
		// Packed data fits in a pointer, so store it directly inside the
		// pointer.
		if len(values) == 1 && values[0].Type().TypeKind() == llvm.IntegerTypeKind {
			// Try to keep this cast in SSA form.
			return c.builder.CreateIntToPtr(values[0], c.i8ptrType, "pack.int")
		}
		// Because packedType is a struct and we have to cast it to a *i8, store
		// it in an alloca first for bitcasting (store+bitcast+load).
		packedAlloc, _, _ = c.createTemporaryAlloca(packedType, "")
	} else {
		// Packed data is bigger than a pointer, so allocate it on the heap.
		sizeValue := llvm.ConstInt(c.uintptrType, size, false)
		packedHeapAlloc = c.createRuntimeCall("alloc", []llvm.Value{sizeValue}, "")
		packedAlloc = c.builder.CreateBitCast(packedHeapAlloc, llvm.PointerType(packedType, 0), "")
	}
	// Store all values in the alloca or heap pointer.
	for i, value := range values {
		indices := []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false),
		}
		gep := c.builder.CreateInBoundsGEP(packedAlloc, indices, "")
		c.builder.CreateStore(value, gep)
	}

	if packedHeapAlloc.IsNil() {
		// Load value (as *i8) from the alloca.
		packedAlloc = c.builder.CreateBitCast(packedAlloc, llvm.PointerType(c.i8ptrType, 0), "")
		result := c.builder.CreateLoad(packedAlloc, "")
		packedPtr := c.builder.CreateBitCast(packedAlloc, c.i8ptrType, "")
		packedSize := llvm.ConstInt(c.ctx.Int64Type(), c.targetData.TypeAllocSize(packedAlloc.Type()), false)
		c.emitLifetimeEnd(packedPtr, packedSize)
		return result
	} else {
		// Get the original heap allocation pointer, which already is an *i8.
		return packedHeapAlloc
	}
}

// emitPointerUnpack extracts a list of values packed using emitPointerPack.
func (c *Compiler) emitPointerUnpack(ptr llvm.Value, valueTypes []llvm.Type) []llvm.Value {
	packedType := c.ctx.StructType(valueTypes, false)

	// Get a correctly-typed pointer to the packed data.
	var packedAlloc, packedRawAlloc llvm.Value
	size := c.targetData.TypeAllocSize(packedType)
	if size == 0 {
		// No data to unpack.
	} else if len(valueTypes) == 1 && valueTypes[0].TypeKind() == llvm.PointerTypeKind {
		// A single pointer is always stored directly.
		return []llvm.Value{c.builder.CreateBitCast(ptr, valueTypes[0], "unpack.ptr")}
	} else if size <= c.targetData.TypeAllocSize(c.i8ptrType) {
		// Packed data stored directly in pointer.
		if len(valueTypes) == 1 && valueTypes[0].TypeKind() == llvm.IntegerTypeKind {
			// Keep this cast in SSA form.
			return []llvm.Value{c.builder.CreatePtrToInt(ptr, valueTypes[0], "unpack.int")}
		}
		// Fallback: load it using an alloca.
		packedRawAlloc, _, _ = c.createTemporaryAlloca(llvm.PointerType(c.i8ptrType, 0), "unpack.raw.alloc")
		packedRawValue := c.builder.CreateBitCast(ptr, llvm.PointerType(c.i8ptrType, 0), "unpack.raw.value")
		c.builder.CreateStore(packedRawValue, packedRawAlloc)
		packedAlloc = c.builder.CreateBitCast(packedRawAlloc, llvm.PointerType(packedType, 0), "unpack.alloc")
	} else {
		// Packed data stored on the heap. Bitcast the passed-in pointer to the
		// correct pointer type.
		packedAlloc = c.builder.CreateBitCast(ptr, llvm.PointerType(packedType, 0), "unpack.raw.ptr")
	}
	// Load each value from the packed data.
	values := make([]llvm.Value, len(valueTypes))
	for i, valueType := range valueTypes {
		if c.targetData.TypeAllocSize(valueType) == 0 {
			// This value has length zero, so there's nothing to load.
			values[i] = c.getZeroValue(valueType)
			continue
		}
		indices := []llvm.Value{
			llvm.ConstInt(c.ctx.Int32Type(), 0, false),
			llvm.ConstInt(c.ctx.Int32Type(), uint64(i), false),
		}
		gep := c.builder.CreateInBoundsGEP(packedAlloc, indices, "")
		values[i] = c.builder.CreateLoad(gep, "")
	}
	if !packedRawAlloc.IsNil() {
		allocPtr := c.builder.CreateBitCast(packedRawAlloc, c.i8ptrType, "")
		allocSize := llvm.ConstInt(c.ctx.Int64Type(), c.targetData.TypeAllocSize(c.uintptrType), false)
		c.emitLifetimeEnd(allocPtr, allocSize)
	}
	return values
}

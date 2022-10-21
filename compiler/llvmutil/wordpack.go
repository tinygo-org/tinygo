package llvmutil

// This file contains utility functions to pack and unpack sets of values. It
// can take in a list of values and tries to store it efficiently in the pointer
// itself if possible and legal.

import (
	"tinygo.org/x/go-llvm"
)

// EmitPointerPack packs the list of values into a single pointer value using
// bitcasts, or else allocates a value on the heap if it cannot be packed in the
// pointer value directly. It returns the pointer with the packed data.
// If the values are all constants, they are be stored in a constant global and deduplicated.
func EmitPointerPack(builder llvm.Builder, mod llvm.Module, prefix string, needsStackObjects bool, values []llvm.Value) llvm.Value {
	ctx := mod.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	i8ptrType := llvm.PointerType(mod.Context().Int8Type(), 0)
	uintptrType := ctx.IntType(targetData.PointerSize() * 8)

	valueTypes := make([]llvm.Type, len(values))
	for i, value := range values {
		valueTypes[i] = value.Type()
	}
	packedType := ctx.StructType(valueTypes, false)

	// Allocate memory for the packed data.
	size := targetData.TypeAllocSize(packedType)
	if size == 0 {
		return llvm.ConstPointerNull(i8ptrType)
	} else if len(values) == 1 && values[0].Type().TypeKind() == llvm.PointerTypeKind {
		return builder.CreateBitCast(values[0], i8ptrType, "pack.ptr")
	} else if size <= targetData.TypeAllocSize(i8ptrType) {
		// Packed data fits in a pointer, so store it directly inside the
		// pointer.
		if len(values) == 1 && values[0].Type().TypeKind() == llvm.IntegerTypeKind {
			// Try to keep this cast in SSA form.
			return builder.CreateIntToPtr(values[0], i8ptrType, "pack.int")
		}

		// Because packedType is a struct and we have to cast it to a *i8, store
		// it in a *i8 alloca first and load the *i8 value from there. This is
		// effectively a bitcast.
		packedAlloc, _, _ := CreateTemporaryAlloca(builder, mod, i8ptrType, "")

		if size < targetData.TypeAllocSize(i8ptrType) {
			// The alloca is bigger than the value that will be stored in it.
			// To avoid having some bits undefined, zero the alloca first.
			// Hopefully this will get optimized away.
			builder.CreateStore(llvm.ConstNull(i8ptrType), packedAlloc)
		}

		// Store all values in the alloca.
		packedAllocCast := builder.CreateBitCast(packedAlloc, llvm.PointerType(packedType, 0), "")
		for i, value := range values {
			indices := []llvm.Value{
				llvm.ConstInt(ctx.Int32Type(), 0, false),
				llvm.ConstInt(ctx.Int32Type(), uint64(i), false),
			}
			gep := builder.CreateInBoundsGEP(packedType, packedAllocCast, indices, "")
			builder.CreateStore(value, gep)
		}

		// Load value (the *i8) from the alloca.
		result := builder.CreateLoad(i8ptrType, packedAlloc, "")

		// End the lifetime of the alloca, to help the optimizer.
		packedPtr := builder.CreateBitCast(packedAlloc, i8ptrType, "")
		packedSize := llvm.ConstInt(ctx.Int64Type(), targetData.TypeAllocSize(packedAlloc.Type()), false)
		EmitLifetimeEnd(builder, mod, packedPtr, packedSize)

		return result
	} else {
		// Check if the values are all constants.
		constant := true
		for _, v := range values {
			if !v.IsConstant() {
				constant = false
				break
			}
		}

		if constant {
			// The data is known at compile time, so store it in a constant global.
			// The global address is marked as unnamed, which allows LLVM to merge duplicates.
			global := llvm.AddGlobal(mod, packedType, prefix+"$pack")
			global.SetInitializer(ctx.ConstStruct(values, false))
			global.SetGlobalConstant(true)
			global.SetUnnamedAddr(true)
			global.SetLinkage(llvm.InternalLinkage)
			return llvm.ConstBitCast(global, i8ptrType)
		}

		// Packed data is bigger than a pointer, so allocate it on the heap.
		sizeValue := llvm.ConstInt(uintptrType, size, false)
		alloc := mod.NamedFunction("runtime.alloc")
		packedHeapAlloc := builder.CreateCall(alloc.GlobalValueType(), alloc, []llvm.Value{
			sizeValue,
			llvm.ConstNull(i8ptrType),
			llvm.Undef(i8ptrType), // unused context parameter
		}, "")
		if needsStackObjects {
			trackPointer := mod.NamedFunction("runtime.trackPointer")
			builder.CreateCall(trackPointer.GlobalValueType(), trackPointer, []llvm.Value{
				packedHeapAlloc,
				llvm.Undef(i8ptrType), // unused context parameter
			}, "")
		}
		packedAlloc := builder.CreateBitCast(packedHeapAlloc, llvm.PointerType(packedType, 0), "")

		// Store all values in the heap pointer.
		for i, value := range values {
			indices := []llvm.Value{
				llvm.ConstInt(ctx.Int32Type(), 0, false),
				llvm.ConstInt(ctx.Int32Type(), uint64(i), false),
			}
			gep := builder.CreateInBoundsGEP(packedType, packedAlloc, indices, "")
			builder.CreateStore(value, gep)
		}

		// Return the original heap allocation pointer, which already is an *i8.
		return packedHeapAlloc
	}
}

// EmitPointerUnpack extracts a list of values packed using EmitPointerPack.
func EmitPointerUnpack(builder llvm.Builder, mod llvm.Module, ptr llvm.Value, valueTypes []llvm.Type) []llvm.Value {
	ctx := mod.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	i8ptrType := llvm.PointerType(mod.Context().Int8Type(), 0)
	uintptrType := ctx.IntType(targetData.PointerSize() * 8)

	packedType := ctx.StructType(valueTypes, false)

	// Get a correctly-typed pointer to the packed data.
	var packedAlloc, packedRawAlloc llvm.Value
	size := targetData.TypeAllocSize(packedType)
	if size == 0 {
		// No data to unpack.
	} else if len(valueTypes) == 1 && valueTypes[0].TypeKind() == llvm.PointerTypeKind {
		// A single pointer is always stored directly.
		return []llvm.Value{builder.CreateBitCast(ptr, valueTypes[0], "unpack.ptr")}
	} else if size <= targetData.TypeAllocSize(i8ptrType) {
		// Packed data stored directly in pointer.
		if len(valueTypes) == 1 && valueTypes[0].TypeKind() == llvm.IntegerTypeKind {
			// Keep this cast in SSA form.
			return []llvm.Value{builder.CreatePtrToInt(ptr, valueTypes[0], "unpack.int")}
		}
		// Fallback: load it using an alloca.
		packedRawAlloc, _, _ = CreateTemporaryAlloca(builder, mod, llvm.PointerType(i8ptrType, 0), "unpack.raw.alloc")
		packedRawValue := builder.CreateBitCast(ptr, llvm.PointerType(i8ptrType, 0), "unpack.raw.value")
		builder.CreateStore(packedRawValue, packedRawAlloc)
		packedAlloc = builder.CreateBitCast(packedRawAlloc, llvm.PointerType(packedType, 0), "unpack.alloc")
	} else {
		// Packed data stored on the heap. Bitcast the passed-in pointer to the
		// correct pointer type.
		packedAlloc = builder.CreateBitCast(ptr, llvm.PointerType(packedType, 0), "unpack.raw.ptr")
	}
	// Load each value from the packed data.
	values := make([]llvm.Value, len(valueTypes))
	for i, valueType := range valueTypes {
		if targetData.TypeAllocSize(valueType) == 0 {
			// This value has length zero, so there's nothing to load.
			values[i] = llvm.ConstNull(valueType)
			continue
		}
		indices := []llvm.Value{
			llvm.ConstInt(ctx.Int32Type(), 0, false),
			llvm.ConstInt(ctx.Int32Type(), uint64(i), false),
		}
		gep := builder.CreateInBoundsGEP(packedType, packedAlloc, indices, "")
		values[i] = builder.CreateLoad(valueType, gep, "")
	}
	if !packedRawAlloc.IsNil() {
		allocPtr := builder.CreateBitCast(packedRawAlloc, i8ptrType, "")
		allocSize := llvm.ConstInt(ctx.Int64Type(), targetData.TypeAllocSize(uintptrType), false)
		EmitLifetimeEnd(builder, mod, allocPtr, allocSize)
	}
	return values
}

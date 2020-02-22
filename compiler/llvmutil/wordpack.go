package llvmutil

// This file contains utility functions to pack and unpack sets of values. It
// can take in a list of values and tries to store it efficiently in the pointer
// itself if possible and legal.

import (
	"github.com/tinygo-org/tinygo/compileopts"
	"tinygo.org/x/go-llvm"
)

// EmitPointerPack packs the list of values into a single pointer value using
// bitcasts, or else allocates a value on the heap if it cannot be packed in the
// pointer value directly. It returns the pointer with the packed data.
// If the values are all constants, they are be stored in a constant global and deduplicated.
func EmitPointerPack(builder llvm.Builder, mod llvm.Module, config *compileopts.Config, values []llvm.Value) llvm.Value {
	ctx := mod.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	i8ptrType := llvm.PointerType(mod.Context().Int8Type(), 0)
	uintptrType := ctx.IntType(llvm.NewTargetData(mod.DataLayout()).PointerSize() * 8)

	valueTypes := make([]llvm.Type, len(values))
	for i, value := range values {
		valueTypes[i] = value.Type()
	}
	packedType := ctx.StructType(valueTypes, false)

	// Allocate memory for the packed data.
	var packedAlloc, packedHeapAlloc llvm.Value
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
		// it in an alloca first for bitcasting (store+bitcast+load).
		packedAlloc, _, _ = CreateTemporaryAlloca(builder, mod, packedType, "")
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
			funcName := builder.GetInsertBlock().Parent().Name()
			global := llvm.AddGlobal(mod, packedType, funcName+"$pack")
			global.SetInitializer(ctx.ConstStruct(values, false))
			global.SetGlobalConstant(true)
			global.SetUnnamedAddr(true)
			global.SetLinkage(llvm.PrivateLinkage)
			return llvm.ConstBitCast(global, i8ptrType)
		}

		// Packed data is bigger than a pointer, so allocate it on the heap.
		sizeValue := llvm.ConstInt(uintptrType, size, false)
		alloc := mod.NamedFunction("runtime.alloc")
		packedHeapAlloc = builder.CreateCall(alloc, []llvm.Value{
			sizeValue,
			llvm.Undef(i8ptrType),            // unused context parameter
			llvm.ConstPointerNull(i8ptrType), // coroutine handle
		}, "")
		if config.NeedsStackObjects() {
			trackPointer := mod.NamedFunction("runtime.trackPointer")
			builder.CreateCall(trackPointer, []llvm.Value{
				packedHeapAlloc,
				llvm.Undef(i8ptrType),            // unused context parameter
				llvm.ConstPointerNull(i8ptrType), // coroutine handle
			}, "")
		}
		packedAlloc = builder.CreateBitCast(packedHeapAlloc, llvm.PointerType(packedType, 0), "")
	}
	// Store all values in the alloca or heap pointer.
	for i, value := range values {
		indices := []llvm.Value{
			llvm.ConstInt(ctx.Int32Type(), 0, false),
			llvm.ConstInt(ctx.Int32Type(), uint64(i), false),
		}
		gep := builder.CreateInBoundsGEP(packedAlloc, indices, "")
		builder.CreateStore(value, gep)
	}

	if packedHeapAlloc.IsNil() {
		// Load value (as *i8) from the alloca.
		packedAlloc = builder.CreateBitCast(packedAlloc, llvm.PointerType(i8ptrType, 0), "")
		result := builder.CreateLoad(packedAlloc, "")
		packedPtr := builder.CreateBitCast(packedAlloc, i8ptrType, "")
		packedSize := llvm.ConstInt(ctx.Int64Type(), targetData.TypeAllocSize(packedAlloc.Type()), false)
		EmitLifetimeEnd(builder, mod, packedPtr, packedSize)
		return result
	} else {
		// Get the original heap allocation pointer, which already is an *i8.
		return packedHeapAlloc
	}
}

// EmitPointerUnpack extracts a list of values packed using EmitPointerPack.
func EmitPointerUnpack(builder llvm.Builder, mod llvm.Module, ptr llvm.Value, valueTypes []llvm.Type) []llvm.Value {
	ctx := mod.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	i8ptrType := llvm.PointerType(mod.Context().Int8Type(), 0)
	uintptrType := ctx.IntType(llvm.NewTargetData(mod.DataLayout()).PointerSize() * 8)

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
		gep := builder.CreateInBoundsGEP(packedAlloc, indices, "")
		values[i] = builder.CreateLoad(gep, "")
	}
	if !packedRawAlloc.IsNil() {
		allocPtr := builder.CreateBitCast(packedRawAlloc, i8ptrType, "")
		allocSize := llvm.ConstInt(ctx.Int64Type(), targetData.TypeAllocSize(uintptrType), false)
		EmitLifetimeEnd(builder, mod, allocPtr, allocSize)
	}
	return values
}

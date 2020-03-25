package compiler

// This file implements functions that do certain safety checks that are
// required by the Go programming language.

import (
	"fmt"
	"go/token"
	"go/types"

	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// createLookupBoundsCheck emits a bounds check before doing a lookup into a
// slice. This is required by the Go language spec: an index out of bounds must
// cause a panic.
func (b *builder) createLookupBoundsCheck(arrayLen, index llvm.Value, indexType types.Type) {
	if b.fn.IsNoBounds() {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	if index.Type().IntTypeWidth() < arrayLen.Type().IntTypeWidth() {
		// Sometimes, the index can be e.g. an uint8 or int8, and we have to
		// correctly extend that type.
		if indexType.Underlying().(*types.Basic).Info()&types.IsUnsigned == 0 {
			index = b.CreateZExt(index, arrayLen.Type(), "")
		} else {
			index = b.CreateSExt(index, arrayLen.Type(), "")
		}
	} else if index.Type().IntTypeWidth() > arrayLen.Type().IntTypeWidth() {
		// The index is bigger than the array length type, so extend it.
		arrayLen = b.CreateZExt(arrayLen, index.Type(), "")
	}

	// Now do the bounds check: index >= arrayLen
	outOfBounds := b.CreateICmp(llvm.IntUGE, index, arrayLen, "")
	b.createRuntimeAssert(outOfBounds, "lookup", "lookupPanic")
}

// createSliceBoundsCheck emits a bounds check before a slicing operation to make
// sure it is within bounds.
//
// This function is both used for slicing a slice (low and high have their
// normal meaning) and for creating a new slice, where 'capacity' means the
// biggest possible slice capacity, 'low' means len and 'high' means cap. The
// logic is the same in both cases.
func (b *builder) createSliceBoundsCheck(capacity, low, high, max llvm.Value, lowType, highType, maxType *types.Basic) {
	if b.fn.IsNoBounds() {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	// Extend the capacity integer to be at least as wide as low and high.
	capacityType := capacity.Type()
	if low.Type().IntTypeWidth() > capacityType.IntTypeWidth() {
		capacityType = low.Type()
	}
	if high.Type().IntTypeWidth() > capacityType.IntTypeWidth() {
		capacityType = high.Type()
	}
	if max.Type().IntTypeWidth() > capacityType.IntTypeWidth() {
		capacityType = max.Type()
	}
	if capacityType != capacity.Type() {
		capacity = b.CreateZExt(capacity, capacityType, "")
	}

	// Extend low and high to be the same size as capacity.
	if low.Type().IntTypeWidth() < capacityType.IntTypeWidth() {
		if lowType.Info()&types.IsUnsigned != 0 {
			low = b.CreateZExt(low, capacityType, "")
		} else {
			low = b.CreateSExt(low, capacityType, "")
		}
	}
	if high.Type().IntTypeWidth() < capacityType.IntTypeWidth() {
		if highType.Info()&types.IsUnsigned != 0 {
			high = b.CreateZExt(high, capacityType, "")
		} else {
			high = b.CreateSExt(high, capacityType, "")
		}
	}
	if max.Type().IntTypeWidth() < capacityType.IntTypeWidth() {
		if maxType.Info()&types.IsUnsigned != 0 {
			max = b.CreateZExt(max, capacityType, "")
		} else {
			max = b.CreateSExt(max, capacityType, "")
		}
	}

	// Now do the bounds check: low > high || high > capacity
	outOfBounds1 := b.CreateICmp(llvm.IntUGT, low, high, "slice.lowhigh")
	outOfBounds2 := b.CreateICmp(llvm.IntUGT, high, max, "slice.highmax")
	outOfBounds3 := b.CreateICmp(llvm.IntUGT, max, capacity, "slice.maxcap")
	outOfBounds := b.CreateOr(outOfBounds1, outOfBounds2, "slice.lowmax")
	outOfBounds = b.CreateOr(outOfBounds, outOfBounds3, "slice.lowcap")
	b.createRuntimeAssert(outOfBounds, "slice", "slicePanic")
}

// createChanBoundsCheck creates a bounds check before creating a new channel to
// check that the value is not too big for runtime.chanMake.
func (b *builder) createChanBoundsCheck(elementSize uint64, bufSize llvm.Value, bufSizeType *types.Basic, pos token.Pos) {
	if b.fn.IsNoBounds() {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	// Check whether the bufSize parameter must be cast to a wider integer for
	// comparison.
	if bufSize.Type().IntTypeWidth() < b.uintptrType.IntTypeWidth() {
		if bufSizeType.Info()&types.IsUnsigned != 0 {
			// Unsigned, so zero-extend to uint type.
			bufSizeType = types.Typ[types.Uint]
			bufSize = b.CreateZExt(bufSize, b.intType, "")
		} else {
			// Signed, so sign-extend to int type.
			bufSizeType = types.Typ[types.Int]
			bufSize = b.CreateSExt(bufSize, b.intType, "")
		}
	}

	// Calculate (^uintptr(0)) >> 1, which is the max value that fits in an
	// uintptr if uintptrs were signed.
	maxBufSize := llvm.ConstLShr(llvm.ConstNot(llvm.ConstInt(b.uintptrType, 0, false)), llvm.ConstInt(b.uintptrType, 1, false))
	if elementSize > maxBufSize.ZExtValue() {
		b.addError(pos, fmt.Sprintf("channel element type is too big (%v bytes)", elementSize))
		return
	}
	// Avoid divide-by-zero.
	if elementSize == 0 {
		elementSize = 1
	}
	// Make the maxBufSize actually the maximum allowed value (in number of
	// elements in the channel buffer).
	maxBufSize = llvm.ConstUDiv(maxBufSize, llvm.ConstInt(b.uintptrType, elementSize, false))

	// Make sure maxBufSize has the same type as bufSize.
	if maxBufSize.Type() != bufSize.Type() {
		maxBufSize = llvm.ConstZExt(maxBufSize, bufSize.Type())
	}

	// Do the check for a too large (or negative) buffer size.
	bufSizeTooBig := b.CreateICmp(llvm.IntUGE, bufSize, maxBufSize, "")
	b.createRuntimeAssert(bufSizeTooBig, "chan", "chanMakePanic")
}

// createNilCheck checks whether the given pointer is nil, and panics if it is.
// It has no effect in well-behaved programs, but makes sure no uncaught nil
// pointer dereferences exist in valid Go code.
func (b *builder) createNilCheck(inst ssa.Value, ptr llvm.Value, blockPrefix string) {
	// Check whether we need to emit this check at all.
	if !ptr.IsAGlobalValue().IsNil() {
		return
	}

	switch inst.(type) {
	case *ssa.IndexAddr:
		// This pointer is the result of an index operation into a slice or
		// array. Such slices/arrays are already bounds checked so the pointer
		// must be a valid (non-nil) pointer. No nil checking is necessary.
		return
	}

	// Compare against nil.
	var isnil llvm.Value
	if ptr.Type().PointerAddressSpace() == 0 {
		// Do the nil check using the isnil builtin, which marks the parameter
		// as nocapture.
		// The reason it has to go through a builtin, is that a regular icmp
		// instruction may capture the pointer in LLVM semantics, see
		// https://reviews.llvm.org/D60047 for details. Pointer capturing
		// unfortunately breaks escape analysis, so we use this trick to let the
		// functionattr pass know that this pointer doesn't really escape.
		ptr = b.CreateBitCast(ptr, b.i8ptrType, "")
		isnil = b.createRuntimeCall("isnil", []llvm.Value{ptr}, "")
	} else {
		// Do the nil check using a regular icmp. This can happen with function
		// pointers on AVR, which don't benefit from escape analysis anyway.
		nilptr := llvm.ConstPointerNull(ptr.Type())
		isnil = b.CreateICmp(llvm.IntEQ, ptr, nilptr, "")
	}

	// Emit the nil check in IR.
	b.createRuntimeAssert(isnil, blockPrefix, "nilPanic")
}

// createRuntimeAssert is a common function to create a new branch on an assert
// bool, calling an assert func if the assert value is true (1).
func (b *builder) createRuntimeAssert(assert llvm.Value, blockPrefix, assertFunc string) {
	// Check whether we can resolve this check at compile time.
	if !assert.IsAConstantInt().IsNil() {
		val := assert.ZExtValue()
		if val == 0 {
			// Everything is constant so the check does not have to be emitted
			// in IR. This avoids emitting some redundant IR.
			return
		}
	}

	faultBlock := b.ctx.AddBasicBlock(b.fn.LLVMFn, blockPrefix+".throw")
	nextBlock := b.ctx.AddBasicBlock(b.fn.LLVMFn, blockPrefix+".next")
	b.blockExits[b.currentBlock] = nextBlock // adjust outgoing block for phi nodes

	// Now branch to the out-of-bounds or the regular block.
	b.CreateCondBr(assert, faultBlock, nextBlock)

	// Fail: the assert triggered so panic.
	b.SetInsertPointAtEnd(faultBlock)
	b.createRuntimeCall(assertFunc, nil, "")
	b.CreateUnreachable()

	// Ok: assert didn't trigger so continue normally.
	b.SetInsertPointAtEnd(nextBlock)
}

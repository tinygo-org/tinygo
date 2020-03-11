package compiler

// This file implements functions that do certain safety checks that are
// required by the Go programming language.

import (
	"fmt"
	"go/token"
	"go/types"

	"tinygo.org/x/go-llvm"
)

// emitLookupBoundsCheck emits a bounds check before doing a lookup into a
// slice. This is required by the Go language spec: an index out of bounds must
// cause a panic.
func (c *Compiler) emitLookupBoundsCheck(frame *Frame, arrayLen, index llvm.Value, indexType types.Type) {
	if frame.fn.IsNoBounds() {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	if index.Type().IntTypeWidth() < arrayLen.Type().IntTypeWidth() {
		// Sometimes, the index can be e.g. an uint8 or int8, and we have to
		// correctly extend that type.
		if indexType.Underlying().(*types.Basic).Info()&types.IsUnsigned == 0 {
			index = c.builder.CreateZExt(index, arrayLen.Type(), "")
		} else {
			index = c.builder.CreateSExt(index, arrayLen.Type(), "")
		}
	} else if index.Type().IntTypeWidth() > arrayLen.Type().IntTypeWidth() {
		// The index is bigger than the array length type, so extend it.
		arrayLen = c.builder.CreateZExt(arrayLen, index.Type(), "")
	}

	// Now do the bounds check: index >= arrayLen
	outOfBounds := c.builder.CreateICmp(llvm.IntUGE, index, arrayLen, "")
	c.createRuntimeAssert(frame, outOfBounds, "lookup", "lookupPanic")
}

// emitSliceBoundsCheck emits a bounds check before a slicing operation to make
// sure it is within bounds.
//
// This function is both used for slicing a slice (low and high have their
// normal meaning) and for creating a new slice, where 'capacity' means the
// biggest possible slice capacity, 'low' means len and 'high' means cap. The
// logic is the same in both cases.
func (c *Compiler) emitSliceBoundsCheck(frame *Frame, capacity, low, high, max llvm.Value, lowType, highType, maxType *types.Basic) {
	if frame.fn.IsNoBounds() {
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
		capacity = c.builder.CreateZExt(capacity, capacityType, "")
	}

	// Extend low and high to be the same size as capacity.
	if low.Type().IntTypeWidth() < capacityType.IntTypeWidth() {
		if lowType.Info()&types.IsUnsigned != 0 {
			low = c.builder.CreateZExt(low, capacityType, "")
		} else {
			low = c.builder.CreateSExt(low, capacityType, "")
		}
	}
	if high.Type().IntTypeWidth() < capacityType.IntTypeWidth() {
		if highType.Info()&types.IsUnsigned != 0 {
			high = c.builder.CreateZExt(high, capacityType, "")
		} else {
			high = c.builder.CreateSExt(high, capacityType, "")
		}
	}
	if max.Type().IntTypeWidth() < capacityType.IntTypeWidth() {
		if maxType.Info()&types.IsUnsigned != 0 {
			max = c.builder.CreateZExt(max, capacityType, "")
		} else {
			max = c.builder.CreateSExt(max, capacityType, "")
		}
	}

	// Now do the bounds check: low > high || high > capacity
	outOfBounds1 := c.builder.CreateICmp(llvm.IntUGT, low, high, "slice.lowhigh")
	outOfBounds2 := c.builder.CreateICmp(llvm.IntUGT, high, max, "slice.highmax")
	outOfBounds3 := c.builder.CreateICmp(llvm.IntUGT, max, capacity, "slice.maxcap")
	outOfBounds := c.builder.CreateOr(outOfBounds1, outOfBounds2, "slice.lowmax")
	outOfBounds = c.builder.CreateOr(outOfBounds, outOfBounds3, "slice.lowcap")
	c.createRuntimeAssert(frame, outOfBounds, "slice", "slicePanic")
}

// emitChanBoundsCheck emits a bounds check before creating a new channel to
// check that the value is not too big for runtime.chanMake.
func (c *Compiler) emitChanBoundsCheck(frame *Frame, elementSize uint64, bufSize llvm.Value, bufSizeType *types.Basic, pos token.Pos) {
	if frame.fn.IsNoBounds() {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	// Check whether the bufSize parameter must be cast to a wider integer for
	// comparison.
	if bufSize.Type().IntTypeWidth() < c.uintptrType.IntTypeWidth() {
		if bufSizeType.Info()&types.IsUnsigned != 0 {
			// Unsigned, so zero-extend to uint type.
			bufSizeType = types.Typ[types.Uint]
			bufSize = c.builder.CreateZExt(bufSize, c.intType, "")
		} else {
			// Signed, so sign-extend to int type.
			bufSizeType = types.Typ[types.Int]
			bufSize = c.builder.CreateSExt(bufSize, c.intType, "")
		}
	}

	// Calculate (^uintptr(0)) >> 1, which is the max value that fits in an
	// uintptr if uintptrs were signed.
	maxBufSize := llvm.ConstLShr(llvm.ConstNot(llvm.ConstInt(c.uintptrType, 0, false)), llvm.ConstInt(c.uintptrType, 1, false))
	if elementSize > maxBufSize.ZExtValue() {
		c.addError(pos, fmt.Sprintf("channel element type is too big (%v bytes)", elementSize))
		return
	}
	// Avoid divide-by-zero.
	if elementSize == 0 {
		elementSize = 1
	}
	// Make the maxBufSize actually the maximum allowed value (in number of
	// elements in the channel buffer).
	maxBufSize = llvm.ConstUDiv(maxBufSize, llvm.ConstInt(c.uintptrType, elementSize, false))

	// Make sure maxBufSize has the same type as bufSize.
	if maxBufSize.Type() != bufSize.Type() {
		maxBufSize = llvm.ConstZExt(maxBufSize, bufSize.Type())
	}

	// Do the check for a too large (or negative) buffer size.
	bufSizeTooBig := c.builder.CreateICmp(llvm.IntUGE, bufSize, maxBufSize, "")
	c.createRuntimeAssert(frame, bufSizeTooBig, "chan", "chanMakePanic")
}

// emitNilCheck checks whether the given pointer is nil, and panics if it is. It
// has no effect in well-behaved programs, but makes sure no uncaught nil
// pointer dereferences exist in valid Go code.
func (c *Compiler) emitNilCheck(frame *Frame, ptr llvm.Value, blockPrefix string) {
	// Check whether we need to emit this check at all.
	if !ptr.IsAGlobalValue().IsNil() {
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
		ptr = c.builder.CreateBitCast(ptr, c.i8ptrType, "")
		isnil = c.createRuntimeCall("isnil", []llvm.Value{ptr}, "")
	} else {
		// Do the nil check using a regular icmp. This can happen with function
		// pointers on AVR, which don't benefit from escape analysis anyway.
		nilptr := llvm.ConstPointerNull(ptr.Type())
		isnil = c.builder.CreateICmp(llvm.IntEQ, ptr, nilptr, "")
	}

	// Emit the nil check in IR.
	c.createRuntimeAssert(frame, isnil, blockPrefix, "nilPanic")
}

// createRuntimeAssert is a common function to create a new branch on an assert
// bool, calling an assert func if the assert value is true (1).
func (c *Compiler) createRuntimeAssert(frame *Frame, assert llvm.Value, blockPrefix, assertFunc string) {
	// Check whether we can resolve this check at compile time.
	if !assert.IsAConstantInt().IsNil() {
		val := assert.ZExtValue()
		if val == 0 {
			// Everything is constant so the check does not have to be emitted
			// in IR. This avoids emitting some redundant IR.
			return
		}
	}

	faultBlock := c.ctx.AddBasicBlock(frame.fn.LLVMFn, blockPrefix+".throw")
	nextBlock := c.ctx.AddBasicBlock(frame.fn.LLVMFn, blockPrefix+".next")
	frame.blockExits[frame.currentBlock] = nextBlock // adjust outgoing block for phi nodes

	// Now branch to the out-of-bounds or the regular block.
	c.builder.CreateCondBr(assert, faultBlock, nextBlock)

	// Fail: the assert triggered so panic.
	c.builder.SetInsertPointAtEnd(faultBlock)
	c.createRuntimeCall(assertFunc, nil, "")
	c.builder.CreateUnreachable()

	// Ok: assert didn't trigger so continue normally.
	c.builder.SetInsertPointAtEnd(nextBlock)
}

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
// The caller should make sure that index is at least as big as arrayLen.
func (b *builder) createLookupBoundsCheck(arrayLen, index llvm.Value) {
	if b.info.nobounds {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	// Extend arrayLen if it's too small.
	if index.Type().IntTypeWidth() > arrayLen.Type().IntTypeWidth() {
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
	if b.info.nobounds {
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
	low = b.extendInteger(low, lowType, capacityType)
	high = b.extendInteger(high, highType, capacityType)
	max = b.extendInteger(max, maxType, capacityType)

	// Now do the bounds check: low > high || high > capacity
	outOfBounds1 := b.CreateICmp(llvm.IntUGT, low, high, "slice.lowhigh")
	outOfBounds2 := b.CreateICmp(llvm.IntUGT, high, max, "slice.highmax")
	outOfBounds3 := b.CreateICmp(llvm.IntUGT, max, capacity, "slice.maxcap")
	outOfBounds := b.CreateOr(outOfBounds1, outOfBounds2, "slice.lowmax")
	outOfBounds = b.CreateOr(outOfBounds, outOfBounds3, "slice.lowcap")
	b.createRuntimeAssert(outOfBounds, "slice", "slicePanic")
}

// createSliceToArrayPointerCheck adds a check for slice-to-array pointer
// conversions. This conversion was added in Go 1.17. For details, see:
// https://tip.golang.org/ref/spec#Conversions_from_slice_to_array_pointer
func (b *builder) createSliceToArrayPointerCheck(sliceLen llvm.Value, arrayLen int64) {
	// From the spec:
	// > If the length of the slice is less than the length of the array, a
	// > run-time panic occurs.
	arrayLenValue := llvm.ConstInt(b.uintptrType, uint64(arrayLen), false)
	isLess := b.CreateICmp(llvm.IntULT, sliceLen, arrayLenValue, "")
	b.createRuntimeAssert(isLess, "slicetoarray", "sliceToArrayPointerPanic")
}

// createUnsafeSliceCheck inserts a runtime check used for unsafe.Slice. This
// function must panic if the ptr/len parameters are invalid.
func (b *builder) createUnsafeSliceCheck(ptr, len llvm.Value, lenType *types.Basic) {
	// From the documentation of unsafe.Slice:
	//   > At run time, if len is negative, or if ptr is nil and len is not
	//   > zero, a run-time panic occurs.
	// However, in practice, it is also necessary to check that the length is
	// not too big that a GEP wouldn't be possible without wrapping the pointer.
	// These two checks (non-negative and not too big) can be merged into one
	// using an unsiged greater than.

	// Make sure the len value is at least as big as a uintptr.
	len = b.extendInteger(len, lenType, b.uintptrType)

	// Determine the maximum slice size, and therefore the maximum value of the
	// len parameter.
	maxSize := b.maxSliceSize(ptr.Type().ElementType())
	maxSizeValue := llvm.ConstInt(len.Type(), maxSize, false)

	// Do the check. By using unsigned greater than for the length check, signed
	// negative values are also checked (which are very large numbers when
	// interpreted as signed values).
	zero := llvm.ConstInt(len.Type(), 0, false)
	lenOutOfBounds := b.CreateICmp(llvm.IntUGT, len, maxSizeValue, "")
	ptrIsNil := b.CreateICmp(llvm.IntEQ, ptr, llvm.ConstNull(ptr.Type()), "")
	lenIsNotZero := b.CreateICmp(llvm.IntNE, len, zero, "")
	assert := b.CreateAnd(ptrIsNil, lenIsNotZero, "")
	assert = b.CreateOr(assert, lenOutOfBounds, "")
	b.createRuntimeAssert(assert, "unsafe.Slice", "unsafeSlicePanic")
}

// createChanBoundsCheck creates a bounds check before creating a new channel to
// check that the value is not too big for runtime.chanMake.
func (b *builder) createChanBoundsCheck(elementSize uint64, bufSize llvm.Value, bufSizeType *types.Basic, pos token.Pos) {
	if b.info.nobounds {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	// Make sure bufSize is at least as big as maxBufSize (an uintptr).
	bufSize = b.extendInteger(bufSize, bufSizeType, b.uintptrType)

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
	maxBufSize = b.CreateUDiv(maxBufSize, llvm.ConstInt(b.uintptrType, elementSize, false), "")

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

	switch inst := inst.(type) {
	case *ssa.Alloc:
		// An alloc is never nil.
		return
	case *ssa.FreeVar:
		// A free variable is allocated in a parent function and is thus never
		// nil.
		return
	case *ssa.IndexAddr:
		// This pointer is the result of an index operation into a slice or
		// array. Such slices/arrays are already bounds checked so the pointer
		// must be a valid (non-nil) pointer. No nil checking is necessary.
		return
	case *ssa.Convert:
		// This is a pointer that comes from a conversion from unsafe.Pointer.
		// Don't do nil checking because this is unsafe code and the code should
		// know what it is doing.
		// Note: all *ssa.Convert instructions that result in a pointer must
		// come from unsafe.Pointer. Testing here for unsafe.Pointer to be sure.
		if inst.X.Type() == types.Typ[types.UnsafePointer] {
			return
		}
	}

	// Compare against nil.
	// We previously used a hack to make sure this wouldn't break escape
	// analysis, but this is not necessary anymore since
	// https://reviews.llvm.org/D60047 has been merged.
	nilptr := llvm.ConstPointerNull(ptr.Type())
	isnil := b.CreateICmp(llvm.IntEQ, ptr, nilptr, "")

	// Emit the nil check in IR.
	b.createRuntimeAssert(isnil, blockPrefix, "nilPanic")
}

// createNegativeShiftCheck creates an assertion that panics if the given shift value is negative.
// This function assumes that the shift value is signed.
func (b *builder) createNegativeShiftCheck(shift llvm.Value) {
	if b.info.nobounds {
		// Function disabled bounds checking - skip shift check.
		return
	}

	// isNegative = shift < 0
	isNegative := b.CreateICmp(llvm.IntSLT, shift, llvm.ConstInt(shift.Type(), 0, false), "")
	b.createRuntimeAssert(isNegative, "shift", "negativeShiftPanic")
}

// createDivideByZeroCheck asserts that y is not zero. If it is, a runtime panic
// will be emitted. This follows the Go specification which says that a divide
// by zero must cause a run time panic.
func (b *builder) createDivideByZeroCheck(y llvm.Value) {
	if b.info.nobounds {
		return
	}

	// isZero = y == 0
	isZero := b.CreateICmp(llvm.IntEQ, y, llvm.ConstInt(y.Type(), 0, false), "")
	b.createRuntimeAssert(isZero, "divbyzero", "divideByZeroPanic")
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

	// Put the fault block at the end of the function and the next block at the
	// current insert position.
	faultBlock := b.ctx.AddBasicBlock(b.llvmFn, blockPrefix+".throw")
	nextBlock := b.insertBasicBlock(blockPrefix + ".next")
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

// extendInteger extends the value to at least targetType using a zero or sign
// extend. The resulting value is not truncated: it may still be bigger than
// targetType.
func (b *builder) extendInteger(value llvm.Value, valueType types.Type, targetType llvm.Type) llvm.Value {
	if value.Type().IntTypeWidth() < targetType.IntTypeWidth() {
		if valueType.Underlying().(*types.Basic).Info()&types.IsUnsigned != 0 {
			// Unsigned, so zero-extend to the target type.
			value = b.CreateZExt(value, targetType, "")
		} else {
			// Signed, so sign-extend to the target type.
			value = b.CreateSExt(value, targetType, "")
		}
	}
	return value
}

package compiler

// This file implements functions that do certain safety checks that are
// required by the Go programming language.

import (
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

	// Sometimes, the index can be e.g. an uint8 or int8, and we have to
	// correctly extend that type.
	if index.Type().IntTypeWidth() < arrayLen.Type().IntTypeWidth() {
		if indexType.(*types.Basic).Info()&types.IsUnsigned == 0 {
			index = c.builder.CreateZExt(index, arrayLen.Type(), "")
		} else {
			index = c.builder.CreateSExt(index, arrayLen.Type(), "")
		}
	}

	// Optimize away trivial cases.
	// LLVM would do this anyway with interprocedural optimizations, but it
	// helps to see cases where bounds check elimination would really help.
	if index.IsConstant() && arrayLen.IsConstant() && !arrayLen.IsUndef() {
		index := index.SExtValue()
		arrayLen := arrayLen.SExtValue()
		if index >= 0 && index < arrayLen {
			return
		}
	}

	if index.Type().IntTypeWidth() > c.intType.IntTypeWidth() {
		// Index is too big for the regular bounds check. Use the one for int64.
		c.createRuntimeCall("lookupBoundsCheckLong", []llvm.Value{arrayLen, index}, "")
	} else {
		c.createRuntimeCall("lookupBoundsCheck", []llvm.Value{arrayLen, index}, "")
	}
}

// emitSliceBoundsCheck emits a bounds check before a slicing operation to make
// sure it is within bounds.
func (c *Compiler) emitSliceBoundsCheck(frame *Frame, capacity, low, high llvm.Value, lowType, highType *types.Basic) {
	if frame.fn.IsNoBounds() {
		// The //go:nobounds pragma was added to the function to avoid bounds
		// checking.
		return
	}

	uintptrWidth := c.uintptrType.IntTypeWidth()
	if low.Type().IntTypeWidth() > uintptrWidth || high.Type().IntTypeWidth() > uintptrWidth {
		if low.Type().IntTypeWidth() < 64 {
			if lowType.Info()&types.IsUnsigned != 0 {
				low = c.builder.CreateZExt(low, c.ctx.Int64Type(), "")
			} else {
				low = c.builder.CreateSExt(low, c.ctx.Int64Type(), "")
			}
		}
		if high.Type().IntTypeWidth() < 64 {
			if highType.Info()&types.IsUnsigned != 0 {
				high = c.builder.CreateZExt(high, c.ctx.Int64Type(), "")
			} else {
				high = c.builder.CreateSExt(high, c.ctx.Int64Type(), "")
			}
		}
		// TODO: 32-bit or even 16-bit slice bounds checks for 8-bit platforms
		c.createRuntimeCall("sliceBoundsCheck64", []llvm.Value{capacity, low, high}, "")
	} else {
		c.createRuntimeCall("sliceBoundsCheck", []llvm.Value{capacity, low, high}, "")
	}
}

// emitNilCheck checks whether the given pointer is nil, and panics if it is. It
// has no effect in well-behaved programs, but makes sure no uncaught nil
// pointer dereferences exist in valid Go code.
func (c *Compiler) emitNilCheck(frame *Frame, ptr llvm.Value, blockPrefix string) {
	// Check whether this is a nil pointer.
	faultBlock := c.ctx.AddBasicBlock(frame.fn.LLVMFn, blockPrefix+".nil")
	nextBlock := c.ctx.AddBasicBlock(frame.fn.LLVMFn, blockPrefix+".next")
	frame.blockExits[frame.currentBlock] = nextBlock // adjust outgoing block for phi nodes

	// Compare against nil.
	nilptr := llvm.ConstPointerNull(ptr.Type())
	isnil := c.builder.CreateICmp(llvm.IntEQ, ptr, nilptr, "")
	c.builder.CreateCondBr(isnil, faultBlock, nextBlock)

	// Fail: this is a nil pointer, exit with a panic.
	c.builder.SetInsertPointAtEnd(faultBlock)
	c.createRuntimeCall("nilpanic", nil, "")
	c.builder.CreateUnreachable()

	// Ok: this is a valid pointer.
	c.builder.SetInsertPointAtEnd(nextBlock)
}

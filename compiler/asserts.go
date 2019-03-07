package compiler

// This file implements functions that do certain safety checks that are
// required by the Go programming language.

import (
	"tinygo.org/x/go-llvm"
)

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

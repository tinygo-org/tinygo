package interp

// This file contains all decoding of LLVM branch instructions, so that changes
// in how LLVM represents them only need to be handled here.

import "tinygo.org/x/go-llvm"

// opBr is the opcode interp uses for every branch instruction, conditional or
// not. The two forms are told apart by their number of operands.
const opBr = llvm.Br

// isBranch returns whether inst is a (conditional or unconditional) br
// instruction.
func isBranch(inst llvm.Value) bool {
	return !inst.IsABranchInst().IsNil()
}

// isCondBranch returns whether inst is a conditional br instruction.
func isCondBranch(inst llvm.Value) bool {
	return isBranch(inst) && inst.SuccessorsCount() == 2
}

// isUncondBranch returns whether inst is an unconditional br instruction.
func isUncondBranch(inst llvm.Value) bool {
	return isBranch(inst) && inst.SuccessorsCount() == 1
}

// branchCondition returns the i1 condition of a conditional br instruction.
func branchCondition(inst llvm.Value) llvm.Value {
	// The bindings lack LLVMGetCondition, so read the operand directly.
	return inst.Operand(0)
}

// branchThen and branchElse return the destination of a conditional br
// instruction when the condition is true or false, respectively. Successors
// are used instead of operands because the successor order is stable while
// the operand order is not (it is reversed in LLVM 22 and older).
func branchThen(inst llvm.Value) llvm.BasicBlock { return inst.Successor(0) }
func branchElse(inst llvm.Value) llvm.BasicBlock { return inst.Successor(1) }

// branchTarget returns the destination of an unconditional br instruction.
func branchTarget(inst llvm.Value) llvm.BasicBlock { return inst.Successor(0) }

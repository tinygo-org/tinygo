package compiler

import (
	"tinygo.org/x/go-llvm"
)

// This file contains helper functions for LLVM that are not exposed in the Go
// bindings.

// Return a list of values (actually, instructions) where this value is used as
// an operand.
func getUses(value llvm.Value) []llvm.Value {
	if value.IsNil() {
		return nil
	}
	var uses []llvm.Value
	use := value.FirstUse()
	for !use.IsNil() {
		uses = append(uses, use.User())
		use = use.NextUse()
	}
	return uses
}

// createEntryBlockAlloca creates a new alloca in the entry block, even though
// the IR builder is located elsewhere. It assumes that the insert point is
// after the last instruction in the current block. Also, it adds lifetime
// information to the IR signalling that the alloca won't be used before this
// point.
//
// This is useful for creating temporary allocas for intrinsics. Don't forget to
// end the lifetime after you're done with it.
func (c *Compiler) createEntryBlockAlloca(t llvm.Type, name string) (alloca, bitcast, size llvm.Value) {
	currentBlock := c.builder.GetInsertBlock()
	c.builder.SetInsertPointBefore(currentBlock.Parent().EntryBasicBlock().FirstInstruction())
	alloca = c.builder.CreateAlloca(t, name)
	c.builder.SetInsertPointAtEnd(currentBlock)
	bitcast = c.builder.CreateBitCast(alloca, c.i8ptrType, name+".bitcast")
	size = llvm.ConstInt(c.ctx.Int64Type(), c.targetData.TypeAllocSize(t), false)
	c.builder.CreateCall(c.getLifetimeStartFunc(), []llvm.Value{size, bitcast}, "")
	return
}

// getLifetimeStartFunc returns the llvm.lifetime.start intrinsic and creates it
// first if it doesn't exist yet.
func (c *Compiler) getLifetimeStartFunc() llvm.Value {
	fn := c.mod.NamedFunction("llvm.lifetime.start.p0i8")
	if fn.IsNil() {
		fnType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.ctx.Int64Type(), c.i8ptrType}, false)
		fn = llvm.AddFunction(c.mod, "llvm.lifetime.start.p0i8", fnType)
	}
	return fn
}

// getLifetimeEndFunc returns the llvm.lifetime.end intrinsic and creates it
// first if it doesn't exist yet.
func (c *Compiler) getLifetimeEndFunc() llvm.Value {
	fn := c.mod.NamedFunction("llvm.lifetime.end.p0i8")
	if fn.IsNil() {
		fnType := llvm.FunctionType(c.ctx.VoidType(), []llvm.Type{c.ctx.Int64Type(), c.i8ptrType}, false)
		fn = llvm.AddFunction(c.mod, "llvm.lifetime.end.p0i8", fnType)
	}
	return fn
}

// splitBasicBlock splits a LLVM basic block into two parts. All instructions
// after afterInst are moved into a new basic block (created right after the
// current one) with the given name.
func (c *Compiler) splitBasicBlock(afterInst llvm.Value, insertAfter llvm.BasicBlock, name string) llvm.BasicBlock {
	oldBlock := afterInst.InstructionParent()
	newBlock := c.ctx.InsertBasicBlock(insertAfter, name)
	var nextInstructions []llvm.Value // values to move

	// Collect to-be-moved instructions.
	inst := afterInst
	for {
		inst = llvm.NextInstruction(inst)
		if inst.IsNil() {
			break
		}
		nextInstructions = append(nextInstructions, inst)
	}

	// Move instructions.
	c.builder.SetInsertPointAtEnd(newBlock)
	for _, inst := range nextInstructions {
		inst.RemoveFromParentAsInstruction()
		c.builder.Insert(inst)
	}

	// Find PHI nodes to update.
	var phiNodes []llvm.Value // PHI nodes to update
	for bb := insertAfter.Parent().FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
		for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
			if inst.IsAPHINode().IsNil() {
				continue
			}
			needsUpdate := false
			incomingCount := inst.IncomingCount()
			for i := 0; i < incomingCount; i++ {
				if inst.IncomingBlock(i) == oldBlock {
					needsUpdate = true
					break
				}
			}
			if !needsUpdate {
				// PHI node has no incoming edge from the old block.
				continue
			}
			phiNodes = append(phiNodes, inst)
		}
	}

	// Update PHI nodes.
	for _, phi := range phiNodes {
		c.builder.SetInsertPointBefore(phi)
		newPhi := c.builder.CreatePHI(phi.Type(), "")
		incomingCount := phi.IncomingCount()
		incomingVals := make([]llvm.Value, incomingCount)
		incomingBlocks := make([]llvm.BasicBlock, incomingCount)
		for i := 0; i < incomingCount; i++ {
			value := phi.IncomingValue(i)
			block := phi.IncomingBlock(i)
			if block == oldBlock {
				block = newBlock
			}
			incomingVals[i] = value
			incomingBlocks[i] = block
		}
		newPhi.AddIncoming(incomingVals, incomingBlocks)
		phi.ReplaceAllUsesWith(newPhi)
		phi.EraseFromParentAsInstruction()
	}

	return newBlock
}

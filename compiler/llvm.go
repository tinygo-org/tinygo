package compiler

import (
	"github.com/tinygo-org/tinygo/compiler/llvmutil"
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

// createTemporaryAlloca creates a new alloca in the entry block and adds
// lifetime start infromation in the IR signalling that the alloca won't be used
// before this point.
//
// This is useful for creating temporary allocas for intrinsics. Don't forget to
// end the lifetime using emitLifetimeEnd after you're done with it.
func (c *Compiler) createTemporaryAlloca(t llvm.Type, name string) (alloca, bitcast, size llvm.Value) {
	return llvmutil.CreateTemporaryAlloca(c.builder, c.mod, t, name)
}

// emitLifetimeEnd signals the end of an (alloca) lifetime by calling the
// llvm.lifetime.end intrinsic. It is commonly used together with
// createTemporaryAlloca.
func (c *Compiler) emitLifetimeEnd(ptr, size llvm.Value) {
	llvmutil.EmitLifetimeEnd(c.builder, c.mod, ptr, size)
}

// emitPointerPack packs the list of values into a single pointer value using
// bitcasts, or else allocates a value on the heap if it cannot be packed in the
// pointer value directly. It returns the pointer with the packed data.
func (c *Compiler) emitPointerPack(values []llvm.Value) llvm.Value {
	return llvmutil.EmitPointerPack(c.builder, c.mod, c.Config, values)
}

// emitPointerUnpack extracts a list of values packed using emitPointerPack.
func (c *Compiler) emitPointerUnpack(ptr llvm.Value, valueTypes []llvm.Type) []llvm.Value {
	return llvmutil.EmitPointerUnpack(c.builder, c.mod, ptr, valueTypes)
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

// makeGlobalArray creates a new LLVM global with the given name and integers as
// contents, and returns the global.
// Note that it is left with the default linkage etc., you should set
// linkage/constant/etc properties yourself.
func (c *Compiler) makeGlobalArray(buf []byte, name string, elementType llvm.Type) llvm.Value {
	globalType := llvm.ArrayType(elementType, len(buf))
	global := llvm.AddGlobal(c.mod, globalType, name)
	value := llvm.Undef(globalType)
	for i := 0; i < len(buf); i++ {
		ch := uint64(buf[i])
		value = llvm.ConstInsertValue(value, llvm.ConstInt(elementType, ch, false), []uint32{uint32(i)})
	}
	global.SetInitializer(value)
	return global
}

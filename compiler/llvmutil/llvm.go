// Package llvmutil contains utility functions used across multiple compiler
// packages. For example, they may be used by both the compiler pacakge and
// transformation packages.
//
// Normally, utility packages are avoided. However, in this case, the utility
// functions are non-trivial and hard to get right. Copying them to multiple
// places would be a big risk if only one of them is updated.
package llvmutil

import (
	"tinygo.org/x/go-llvm"
)

// CreateEntryBlockAlloca creates a new alloca in the entry block, even though
// the IR builder is located elsewhere. It assumes that the insert point is
// at the end of the current block.
func CreateEntryBlockAlloca(builder llvm.Builder, t llvm.Type, name string) llvm.Value {
	currentBlock := builder.GetInsertBlock()
	entryBlock := currentBlock.Parent().EntryBasicBlock()
	if entryBlock.FirstInstruction().IsNil() {
		builder.SetInsertPointAtEnd(entryBlock)
	} else {
		builder.SetInsertPointBefore(entryBlock.FirstInstruction())
	}
	alloca := builder.CreateAlloca(t, name)
	builder.SetInsertPointAtEnd(currentBlock)
	return alloca
}

// CreateTemporaryAlloca creates a new alloca in the entry block and adds
// lifetime start infromation in the IR signalling that the alloca won't be used
// before this point.
//
// This is useful for creating temporary allocas for intrinsics. Don't forget to
// end the lifetime using emitLifetimeEnd after you're done with it.
func CreateTemporaryAlloca(builder llvm.Builder, mod llvm.Module, t llvm.Type, name string) (alloca, size llvm.Value) {
	ctx := t.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()
	alloca = CreateEntryBlockAlloca(builder, t, name)
	size = llvm.ConstInt(ctx.Int64Type(), targetData.TypeAllocSize(t), false)
	fnType, fn := getLifetimeStartFunc(mod)
	builder.CreateCall(fnType, fn, []llvm.Value{size, alloca}, "")
	return
}

// CreateInstructionAlloca creates an alloca in the entry block, and places lifetime control intrinsics around the instruction
func CreateInstructionAlloca(builder llvm.Builder, mod llvm.Module, t llvm.Type, inst llvm.Value, name string) llvm.Value {
	ctx := mod.Context()
	targetData := llvm.NewTargetData(mod.DataLayout())
	defer targetData.Dispose()

	alloca := CreateEntryBlockAlloca(builder, t, name)
	builder.SetInsertPointBefore(inst)
	size := llvm.ConstInt(ctx.Int64Type(), targetData.TypeAllocSize(t), false)
	fnType, fn := getLifetimeStartFunc(mod)
	builder.CreateCall(fnType, fn, []llvm.Value{size, alloca}, "")
	if next := llvm.NextInstruction(inst); !next.IsNil() {
		builder.SetInsertPointBefore(next)
	} else {
		builder.SetInsertPointAtEnd(inst.InstructionParent())
	}
	fnType, fn = getLifetimeEndFunc(mod)
	builder.CreateCall(fnType, fn, []llvm.Value{size, alloca}, "")
	return alloca
}

// EmitLifetimeEnd signals the end of an (alloca) lifetime by calling the
// llvm.lifetime.end intrinsic. It is commonly used together with
// createTemporaryAlloca.
func EmitLifetimeEnd(builder llvm.Builder, mod llvm.Module, ptr, size llvm.Value) {
	fnType, fn := getLifetimeEndFunc(mod)
	builder.CreateCall(fnType, fn, []llvm.Value{size, ptr}, "")
}

// getLifetimeStartFunc returns the llvm.lifetime.start intrinsic and creates it
// first if it doesn't exist yet.
func getLifetimeStartFunc(mod llvm.Module) (llvm.Type, llvm.Value) {
	fnName := "llvm.lifetime.start.p0"
	fn := mod.NamedFunction(fnName)
	ctx := mod.Context()
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	fnType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ctx.Int64Type(), ptrType}, false)
	if fn.IsNil() {
		fn = llvm.AddFunction(mod, fnName, fnType)
	}
	return fnType, fn
}

// getLifetimeEndFunc returns the llvm.lifetime.end intrinsic and creates it
// first if it doesn't exist yet.
func getLifetimeEndFunc(mod llvm.Module) (llvm.Type, llvm.Value) {
	fnName := "llvm.lifetime.end.p0"
	fn := mod.NamedFunction(fnName)
	ctx := mod.Context()
	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	fnType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ctx.Int64Type(), ptrType}, false)
	if fn.IsNil() {
		fn = llvm.AddFunction(mod, fnName, fnType)
	}
	return fnType, fn
}

// SplitBasicBlock splits a LLVM basic block into two parts. All instructions
// after afterInst are moved into a new basic block (created right after the
// current one) with the given name.
func SplitBasicBlock(builder llvm.Builder, afterInst llvm.Value, insertAfter llvm.BasicBlock, name string) llvm.BasicBlock {
	oldBlock := afterInst.InstructionParent()
	newBlock := afterInst.Type().Context().InsertBasicBlock(insertAfter, name)
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
	builder.SetInsertPointAtEnd(newBlock)
	for _, inst := range nextInstructions {
		inst.RemoveFromParentAsInstruction()
		builder.Insert(inst)
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
		builder.SetInsertPointBefore(phi)
		newPhi := builder.CreatePHI(phi.Type(), "")
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

// Append the given values to a global array like llvm.used. The global might
// not exist yet. The values can be any pointer type, they will be cast to i8*.
func AppendToGlobal(mod llvm.Module, globalName string, values ...llvm.Value) {
	// Read the existing values in the llvm.used array (if it exists).
	var usedValues []llvm.Value
	if used := mod.NamedGlobal(globalName); !used.IsNil() {
		builder := mod.Context().NewBuilder()
		defer builder.Dispose()
		usedInitializer := used.Initializer()
		num := usedInitializer.Type().ArrayLength()
		for i := 0; i < num; i++ {
			usedValues = append(usedValues, builder.CreateExtractValue(usedInitializer, i, ""))
		}
		used.EraseFromParentAsGlobal()
	}

	// Add the new values.
	ptrType := llvm.PointerType(mod.Context().Int8Type(), 0)
	for _, value := range values {
		// Note: the bitcast is necessary to cast AVR function pointers to
		// address space 0 pointer types.
		usedValues = append(usedValues, llvm.ConstPointerCast(value, ptrType))
	}

	// Create a new array (with the old and new values).
	usedInitializer := llvm.ConstArray(ptrType, usedValues)
	used := llvm.AddGlobal(mod, usedInitializer.Type(), globalName)
	used.SetInitializer(usedInitializer)
	used.SetLinkage(llvm.AppendingLinkage)
}

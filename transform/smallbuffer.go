package transform

import (
	"tinygo.org/x/go-llvm"
)

// smallBufferSize is the same buffer size that the Go compiler uses.
const smallBufferSize = 32

// smallBufferBranch holds the blocks. The caller fills small and ends it with a
// branch to cont.
type smallBufferBranch struct {
	buffer            llvm.Value      // the stack buffer
	small, heap, cont llvm.BasicBlock // the paths for sizes that fit and that do not, and the join

	// replaced maps PHIs that were recreated to their replacements. This
	// happens when the block loops back to itself or to a block with PHIs.
	replaced map[llvm.Value]llvm.Value
}

// resolve returns the current value for v, which may have been a PHI that was
// replaced while splitting the block.
func (b smallBufferBranch) resolve(v llvm.Value) llvm.Value {
	return resolveValue(b.replaced, v)
}

// splitSmallBufferBlock branches to small when size fits in a stack buffer and
// to heap, which holds call, when it does not. Both paths join in cont.
func splitSmallBufferBlock(builder llvm.Builder, call, size llvm.Value, align int, prefix string) smallBufferBranch {
	ctx := call.Type().Context()
	block := call.InstructionParent()
	fn := block.Parent()

	// A static alloca in the entry block takes space in the frame once.
	builder.SetInsertPointBefore(fn.EntryBasicBlock().FirstInstruction())
	bufferType := llvm.ArrayType(ctx.Int8Type(), smallBufferSize)
	buffer := builder.CreateAlloca(bufferType, prefix+".buffer")
	buffer.SetAlignment(align)

	small := ctx.AddBasicBlock(fn, prefix+".small")
	heap := ctx.AddBasicBlock(fn, prefix+".heap")
	cont := ctx.AddBasicBlock(fn, prefix+".cont")

	// Everything after the call moves to cont, so the PHIs of its successors
	// must name the new block.
	var rest []llvm.Value
	for inst := llvm.NextInstruction(call); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
		rest = append(rest, inst)
	}
	for _, inst := range rest {
		inst.RemoveFromParentAsInstruction()
		builder.SetInsertPointAtEnd(cont)
		builder.Insert(inst)
	}
	replaced := make(map[llvm.Value]llvm.Value)
	term := cont.LastInstruction()
	for i := 0; i < term.SuccessorsCount(); i++ {
		retargetPHIs(builder, term.Successor(i), block, cont, replaced)
	}
	size = resolveValue(replaced, size)

	builder.SetInsertPointAtEnd(block)
	fits := builder.CreateICmp(llvm.IntULE, size, llvm.ConstInt(size.Type(), smallBufferSize, false), prefix+".fits")
	builder.CreateCondBr(fits, small, heap)

	call.RemoveFromParentAsInstruction()
	builder.SetInsertPointAtEnd(heap)
	builder.Insert(call)
	builder.CreateBr(cont)

	return smallBufferBranch{buffer: buffer, small: small, heap: heap, cont: cont, replaced: replaced}
}

func resolveValue(replaced map[llvm.Value]llvm.Value, v llvm.Value) llvm.Value {
	if r, ok := replaced[v]; ok {
		return r
	}
	return v
}

// retargetPHIs makes the PHIs in block take the values that came from oldPred
// from newPred instead. The PHIs are recreated and recorded in replaced.
func retargetPHIs(builder llvm.Builder, block, oldPred, newPred llvm.BasicBlock, replaced map[llvm.Value]llvm.Value) {
	var phis []llvm.Value
	for inst := block.FirstInstruction(); !inst.IsNil() && !inst.IsAPHINode().IsNil(); inst = llvm.NextInstruction(inst) {
		phis = append(phis, inst)
	}
	for _, phi := range phis {
		var values []llvm.Value
		var blocks []llvm.BasicBlock
		changed := false
		for i := 0; i < phi.IncomingCount(); i++ {
			values = append(values, phi.IncomingValue(i))
			incoming := phi.IncomingBlock(i)
			if incoming == oldPred {
				incoming = newPred
				changed = true
			}
			blocks = append(blocks, incoming)
		}
		if !changed {
			continue
		}
		builder.SetInsertPointBefore(phi)
		replacement := builder.CreatePHI(phi.Type(), "")
		replacement.AddIncoming(values, blocks)
		phi.ReplaceAllUsesWith(replacement)
		replaced[phi] = replacement
		phi.EraseFromParentAsInstruction()
	}
}

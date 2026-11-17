package transform

import (
	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

// addSmallBufferLifetimes adds lifetime markers so LLVM can share one stack slot
// between buffers that are not live together. It adds none for unknown uses.
func addSmallBufferLifetimes(mod llvm.Module, builder llvm.Builder, branch smallBufferBranch) {
	// Collect all values that can point into the buffer.
	derived := map[llvm.Value]bool{branch.buffer: true}
	work := []llvm.Value{branch.buffer}
	for len(work) > 0 {
		v := work[len(work)-1]
		work = work[:len(work)-1]
		for _, use := range getUses(v) {
			if use.IsAInstruction().IsNil() {
				return
			}
			var follow bool
			switch use.InstructionOpcode() {
			case llvm.GetElementPtr, llvm.PHI, llvm.Select, llvm.InsertValue, llvm.BitCast:
				follow = true
			case llvm.ExtractValue:
				follow = typeHasPointers(use.Type())
			case llvm.Load, llvm.ICmp:
				// These do not derive a new pointer.
			case llvm.Store:
				if use.Operand(0) == v {
					// The pointer lives on in memory.
					return
				}
			case llvm.Call:
				// A call can return an alias of its argument.
				follow = typeHasPointers(use.Type())
			default:
				return
			}
			if follow && !derived[use] {
				derived[use] = true
				work = append(work, use)
			}
		}
	}

	fn := branch.small.Parent()
	preds := map[llvm.BasicBlock][]llvm.BasicBlock{}
	var blocks []llvm.BasicBlock
	for b := range llvmutil.BasicBlocks(fn) {
		blocks = append(blocks, b)
		term := b.LastInstruction()
		for i := 0; i < term.SuccessorsCount(); i++ {
			s := term.Successor(i)
			preds[s] = append(preds[s], b)
		}
	}

	// The buffer is considered defined at the start of the small block.
	defBlock := func(v llvm.Value) llvm.BasicBlock {
		if v == branch.buffer {
			return branch.small
		}
		return v.InstructionParent()
	}

	// Standard SSA liveness, unioned over all derived values.
	liveIn := map[llvm.BasicBlock]bool{}
	liveOut := map[llvm.BasicBlock]bool{}
	// A block can be reached from values defined in different blocks. Each
	// walk stops at its own definition, so a block seen for one definition
	// must still be walked for another. Otherwise the result would depend on
	// the order in which the derived values are visited.
	type visit struct{ b, def llvm.BasicBlock }
	visited := map[visit]bool{}
	var markLiveIn func(b, def llvm.BasicBlock)
	markLiveOut := func(b, def llvm.BasicBlock) {
		liveOut[b] = true
		if b != def {
			markLiveIn(b, def)
		}
	}
	markLiveIn = func(b, def llvm.BasicBlock) {
		if visited[visit{b, def}] {
			return
		}
		visited[visit{b, def}] = true
		liveIn[b] = true
		for _, p := range preds[b] {
			markLiveOut(p, def)
		}
	}
	// The last instruction in each block that uses or defines a derived value.
	lastUse := map[llvm.BasicBlock]llvm.Value{}
	note := func(inst llvm.Value) {
		b := inst.InstructionParent()
		last := lastUse[b]
		if last.IsNil() || instructionBefore(last, inst) {
			lastUse[b] = inst
		}
	}
	for v := range derived {
		def := defBlock(v)
		if v != branch.buffer {
			note(v)
		}
		for _, use := range getUses(v) {
			ub := use.InstructionParent()
			if use.InstructionOpcode() == llvm.PHI {
				// Used at the end of the incoming block.
				for i := 0; i < use.IncomingCount(); i++ {
					if use.IncomingValue(i) == v {
						markLiveOut(use.IncomingBlock(i), def)
					}
				}
				continue
			}
			note(use)
			if ub != def {
				markLiveIn(ub, def)
			}
		}
	}
	// A derived phi is defined at the start of its block.
	for v := range derived {
		if v != branch.buffer && v.InstructionOpcode() == llvm.PHI {
			liveIn[v.InstructionParent()] = true
		}
	}

	size := llvm.ConstInt(mod.Context().Int64Type(), smallBufferSize, false)
	buffer := branch.buffer

	builder.SetInsertPointBefore(branch.small.FirstInstruction())
	llvmutil.EmitLifetimeStart(builder, mod, buffer, size)

	for _, b := range blocks {
		if !liveIn[b] && b != branch.small {
			continue
		}
		if liveOut[b] {
			// Only edges to a block with one predecessor can take the marker.
			term := b.LastInstruction()
			for i := 0; i < term.SuccessorsCount(); i++ {
				s := term.Successor(i)
				if liveIn[s] || len(preds[s]) != 1 || s == b {
					continue
				}
				builder.SetInsertPointBefore(firstNonPHI(s))
				llvmutil.EmitLifetimeEnd(builder, mod, buffer, size)
			}
			continue
		}
		// The buffer dies in this block, after its last use.
		last := lastUse[b]
		var next llvm.Value
		switch {
		case last.IsNil():
			next = firstNonPHI(b)
		case !llvm.NextInstruction(last).IsNil():
			next = llvm.NextInstruction(last)
			if !next.IsAPHINode().IsNil() {
				next = firstNonPHI(b)
			}
		default:
			continue // the last use is the terminator
		}
		builder.SetInsertPointBefore(next)
		llvmutil.EmitLifetimeEnd(builder, mod, buffer, size)
	}
}

// firstNonPHI returns the first instruction in the block that is not a PHI.
func firstNonPHI(b llvm.BasicBlock) llvm.Value {
	for inst := range llvmutil.Instructions(b) {
		if inst.IsAPHINode().IsNil() {
			return inst
		}
	}
	return llvm.Value{}
}

// instructionBefore reports whether a comes before b in the same block.
func instructionBefore(a, b llvm.Value) bool {
	seen := false
	for inst := range llvmutil.Instructions(a.InstructionParent()) {
		if inst == b {
			return seen
		}
		seen = seen || inst == a
	}
	return false
}

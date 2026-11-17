package transform

import (
	"strings"

	"tinygo.org/x/go-llvm"
)

// useSmallBufferForAlloc gives a non-escaping allocation of unknown size, like
// make([]byte, n), a stack buffer when it fits. It reports whether it handled it.
func useSmallBufferForAlloc(builder llvm.Builder, alloc llvm.Value, maxAlign int64) bool {
	if alloc.CalledValue().Name() != "runtime.alloc" {
		return false
	}
	if strings.HasPrefix(alloc.InstructionParent().AsValue().Name(), "alloc.heap") {
		// The first run of OptimizeAllocs rewrote this call already.
		return true
	}
	if !valueEscapesAt(alloc).IsNil() {
		return false
	}

	align := int(maxAlign)
	if attr := alloc.GetCallSiteEnumAttribute(0, llvm.AttributeKindID("align")); !attr.IsNil() {
		align = int(attr.GetEnumValue())
	}
	branch := splitSmallBufferBlock(builder, alloc, alloc.Operand(0), align, "alloc")

	// The memory must be zeroed.
	builder.SetInsertPointAtEnd(branch.small)
	zero := builder.CreateStore(llvm.ConstNull(branch.buffer.AllocatedType()), branch.buffer)
	zero.SetAlignment(align)
	builder.CreateBr(branch.cont)

	builder.SetInsertPointBefore(branch.cont.FirstInstruction())
	phi := builder.CreatePHI(alloc.Type(), "alloc.ptr")
	alloc.ReplaceAllUsesWith(phi)
	phi.AddIncoming([]llvm.Value{branch.buffer, alloc}, []llvm.BasicBlock{branch.small, branch.heap})
	return true
}

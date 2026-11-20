package transform

import (
	"strconv"

	"tinygo.org/x/go-llvm"
)

// OptimizeStringFromBytesStack copies string(b) to a 32 byte stack buffer when
// the string does not escape and fits. Unlike OptimizeStringFromBytes it copies.
func OptimizeStringFromBytesStack(mod llvm.Module) {
	stringFromBytes := mod.NamedFunction("runtime.stringFromBytes")
	if stringFromBytes.IsNil() {
		return
	}

	builder := mod.Context().NewBuilder()
	defer builder.Dispose()

	for _, call := range getUses(stringFromBytes) {
		if call.IsACallInst().IsNil() || call.CalledValue() != stringFromBytes {
			continue
		}
		if canCopyStringToStack(call) {
			copyStringToStackWhenSmall(mod, builder, call)
		}
	}
}

// canCopyStringToStack returns whether the result is only used through its
// pointer and length, and the pointer does not escape.
func canCopyStringToStack(call llvm.Value) bool {
	uses := getUses(call)
	if len(uses) == 0 {
		return false
	}
	for _, use := range uses {
		if use.IsAExtractValueInst().IsNil() {
			return false
		}
	}
	return valueEscapesAt(call).IsNil()
}

// copyStringToStackWhenSmall copies the bytes to the stack buffer when the
// length fits, and otherwise makes the original call.
func copyStringToStackWhenSmall(mod llvm.Module, builder llvm.Builder, call llvm.Value) {
	ptr, length := call.Operand(0), call.Operand(1)

	var ptrUses, lenUses []llvm.Value
	for _, use := range getUses(call) {
		if use.Type().TypeKind() == llvm.PointerTypeKind {
			ptrUses = append(ptrUses, use)
		} else {
			lenUses = append(lenUses, use)
		}
	}

	branch := splitSmallBufferBlock(builder, call, length, 1, "string")

	builder.SetInsertPointAtEnd(branch.small)
	createMemcpy(mod, builder, branch.buffer, branch.resolve(ptr), branch.resolve(length))
	builder.CreateBr(branch.cont)

	builder.SetInsertPointBefore(branch.heap.LastInstruction())
	heapPtr := builder.CreateExtractValue(call, 0, "string.heap.ptr")

	builder.SetInsertPointBefore(branch.cont.FirstInstruction())
	phi := builder.CreatePHI(branch.resolve(ptr).Type(), "string.ptr")
	phi.AddIncoming([]llvm.Value{branch.buffer, heapPtr}, []llvm.BasicBlock{branch.small, branch.heap})
	for _, use := range ptrUses {
		use.ReplaceAllUsesWith(phi)
		use.EraseFromParentAsInstruction()
	}
	for _, use := range lenUses {
		use.ReplaceAllUsesWith(branch.resolve(length))
		use.EraseFromParentAsInstruction()
	}
	addSmallBufferLifetimes(mod, builder, branch)
}

// createMemcpy calls llvm.memcpy for a non-volatile copy.
func createMemcpy(mod llvm.Module, builder llvm.Builder, dst, src, length llvm.Value) {
	ctx := mod.Context()
	ptrType := dst.Type()
	lenType := length.Type()
	name := "llvm.memcpy.p0.p0.i" + strconv.Itoa(lenType.IntTypeWidth())
	fnType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ptrType, ptrType, lenType, ctx.Int1Type()}, false)
	fn := mod.NamedFunction(name)
	if fn.IsNil() {
		fn = llvm.AddFunction(mod, name, fnType)
	}
	builder.CreateCall(fnType, fn, []llvm.Value{dst, src, length, llvm.ConstInt(ctx.Int1Type(), 0, false)}, "")
}

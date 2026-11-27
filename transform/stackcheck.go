package transform

import (
	"slices"
	"strings"

	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

const (
	stackLimitName    = "internal/task.stackLimit"
	stackOverflowName = "internal/task.stackOverflow"
)

// KeepStackCheckRuntime makes the stack limit and the overflow handler
// external so they survive until AddStackLimitChecks makes them internal again.
func KeepStackCheckRuntime(mod llvm.Module) {
	if global := mod.NamedGlobal(stackLimitName); !global.IsNil() {
		global.SetLinkage(llvm.ExternalLinkage)
	}
	if fn := mod.NamedFunction(stackOverflowName); !fn.IsNil() && !fn.IsDeclaration() {
		fn.SetLinkage(llvm.ExternalLinkage)
	}
}

// AddStackLimitChecks compares the stack pointer with the stack limit at the
// start of every function that has a stack frame.
func AddStackLimitChecks(mod llvm.Module) {
	limit := mod.NamedGlobal(stackLimitName)
	handler := mod.NamedFunction(stackOverflowName)
	if limit.IsNil() || handler.IsNil() || handler.IsDeclaration() {
		return
	}
	defer limit.SetLinkage(llvm.InternalLinkage)
	defer handler.SetLinkage(llvm.InternalLinkage)
	ctx := mod.Context()
	handler.AddFunctionAttr(ctx.CreateEnumAttribute(llvm.AttributeKindID("cold"), 0))
	handler.AddFunctionAttr(ctx.CreateEnumAttribute(llvm.AttributeKindID("noinline"), 0))

	ptrType := llvm.PointerType(ctx.Int8Type(), 0)
	stacksaveName := "llvm.stacksave.p0"
	if llvm.VersionMajor < 18 {
		stacksaveName = "llvm.stacksave"
	}
	stacksave := mod.NamedFunction(stacksaveName)
	if stacksave.IsNil() {
		stacksave = llvm.AddFunction(mod, stacksaveName, llvm.FunctionType(ptrType, nil, false))
	}
	limitType := limit.GlobalValueType()

	builder := ctx.NewBuilder()
	defer builder.Dispose()

	for fn := range llvmutil.Functions(mod) {
		if fn.IsDeclaration() || fn == handler || !hasStackFrame(fn) {
			continue
		}
		// Move the static allocas into a new entry block that does the check.
		entry := fn.EntryBasicBlock()
		checkBlock := ctx.InsertBasicBlock(entry, "stack.check")
		failBlock := ctx.AddBasicBlock(fn, "stack.fail")
		builder.SetInsertPointAtEnd(checkBlock)
		if sub := fn.Subprogram(); sub.C != nil {
			builder.SetCurrentDebugLocation(sub.SubprogramLine(), 0, sub, llvm.Metadata{})
		} else {
			builder.SetCurrentDebugLocation(0, 0, llvm.Metadata{}, llvm.Metadata{})
		}
		var allocas []llvm.Value
		for inst := range llvmutil.Instructions(entry) {
			if !inst.IsAAllocaInst().IsNil() {
				allocas = append(allocas, inst)
			}
		}
		for _, inst := range allocas {
			inst.RemoveFromParentAsInstruction()
			builder.Insert(inst)
		}
		sp := builder.CreateCall(stacksave.GlobalValueType(), stacksave, nil, "stack.sp")
		spInt := builder.CreatePtrToInt(sp, limitType, "")
		lim := builder.CreateLoad(limitType, limit, "stack.limit")
		overflow := builder.CreateICmp(llvm.IntULT, spInt, lim, "stack.overflow")
		builder.CreateCondBr(overflow, failBlock, entry)

		builder.SetInsertPointAtEnd(failBlock)
		params := make([]llvm.Value, handler.ParamsCount())
		for i := range params {
			params[i] = llvm.Undef(handler.Param(i).Type())
		}
		call := builder.CreateCall(handler.GlobalValueType(), handler, params, "")
		call.AddCallSiteAttribute(-1, ctx.CreateEnumAttribute(llvm.AttributeKindID("cold"), 0))
		builder.CreateUnreachable()
	}
}

// hasStackFrame returns whether fn has allocas, which make a frame in linear
// memory on wasm. It removes dead allocas first.
func hasStackFrame(fn llvm.Value) bool {
	var unused []llvm.Value
	found := false
	for bb := range llvmutil.BasicBlocks(fn) {
		for inst := range llvmutil.Instructions(bb) {
			if inst.IsAAllocaInst().IsNil() {
				continue
			}
			if isDeadAlloca(inst) {
				unused = append(unused, inst)
			} else {
				found = true
			}
		}
	}
	for _, inst := range unused {
		for _, use := range slices.Collect(llvmutil.Uses(inst)) {
			use.EraseFromParentAsInstruction()
		}
		inst.EraseFromParentAsInstruction()
	}
	return found
}

// isDeadAlloca returns whether alloca is only written to and marked with
// lifetime markers, so that it can be removed.
func isDeadAlloca(alloca llvm.Value) bool {
	for use := range llvmutil.Uses(alloca) {
		if !use.IsAStoreInst().IsNil() {
			if use.Operand(1) != alloca || use.Operand(0) == alloca || use.IsVolatile() {
				return false
			}
			continue
		}
		if use.IsACallInst().IsNil() || !strings.HasPrefix(use.CalledValue().Name(), "llvm.lifetime.") {
			return false
		}
	}
	return true
}

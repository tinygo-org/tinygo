package transform

import "tinygo.org/x/go-llvm"

// AddUnwindAssumptions records that normal Go calls only begin while the
// transient Asyncify unwind signal is clear. Asyncify panic unwinding is
// synchronous and cannot be interrupted by another Go entry. Checks are not
// emitted while deferred calls run, and landing pads clear the signal before
// invoking deferred code. LLVM's alias analysis can then remove checks around
// calls that do not modify the signal.
func AddUnwindAssumptions(mod llvm.Module) bool {
	signal := mod.NamedGlobal("runtime.unwindPendingSignal")
	if signal.IsNil() ||
		signal.GlobalValueType().TypeKind() != llvm.IntegerTypeKind ||
		signal.GlobalValueType().IntTypeWidth() != 1 {
		return false
	}
	unwind := mod.NamedFunction("runtime.unwindPending")
	if unwind.IsNil() {
		return false
	}

	functions := make(map[llvm.Value]struct{})
	// Suspension-safe catchers call unwindPending indirectly and are
	// intentionally absent from this set.
	for _, call := range getUses(unwind) {
		if call.IsACallInst().IsNil() || call.CalledValue() != unwind {
			continue
		}
		functions[call.InstructionParent().Parent()] = struct{}{}
	}
	if len(functions) == 0 {
		return false
	}

	ctx := mod.Context()
	assumeType := llvm.FunctionType(ctx.VoidType(), []llvm.Type{ctx.Int1Type()}, false)
	assume := mod.NamedFunction("llvm.assume")
	if assume.IsNil() {
		assume = llvm.AddFunction(mod, "llvm.assume", assumeType)
	}

	builder := ctx.NewBuilder()
	defer builder.Dispose()
	for fn := range functions {
		first := fn.EntryBasicBlock().FirstInstruction()
		builder.SetInsertPointBefore(first)
		unwinding := builder.CreateLoad(ctx.Int1Type(), signal, "unwind.entry")
		notUnwinding := builder.CreateNot(unwinding, "")
		builder.CreateCall(assumeType, assume, []llvm.Value{notUnwinding}, "")
	}
	return true
}

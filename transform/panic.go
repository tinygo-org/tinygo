package transform

import (
	"tinygo.org/x/go-llvm"
)

// ReplacePanicsWithTrap replaces each call to panic (or similar functions) with
// calls to llvm.trap, to reduce code size. This is the -panic=trap command-line
// option.
func ReplacePanicsWithTrap(mod llvm.Module) {
	ctx := mod.Context()
	builder := ctx.NewBuilder()
	defer builder.Dispose()

	trap := mod.NamedFunction("llvm.trap")
	if trap.IsNil() {
		trapType := llvm.FunctionType(ctx.VoidType(), nil, false)
		trap = llvm.AddFunction(mod, "llvm.trap", trapType)
	}
	for _, name := range []string{"runtime._panic", "runtime.runtimePanic"} {
		fn := mod.NamedFunction(name)
		if fn.IsNil() {
			continue
		}
		for _, use := range getUses(fn) {
			if use.IsACallInst().IsNil() || use.CalledValue() != fn {
				panic("expected use of a panic function to be a call")
			}
			builder.SetInsertPointBefore(use)
			builder.CreateCall(trap.GlobalValueType(), trap, nil, "")
		}
	}
}

package transform

import (
	"fmt"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// CheckNoRuntime rejects marked functions that survive dead-code removal.
// Call after interp and before any inlining (builder.optimizeProgram).
func CheckNoRuntime(mod llvm.Module) []error {
	hasMarked := false
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if !fn.GetStringAttributeAtIndex(-1, "tinygo-noruntime").IsNil() {
			hasMarked = true
			break
		}
	}
	if !hasMarked {
		return nil
	}

	// Ignore temporary ABI roots, then restore surviving roots for the optimizer.
	// See compiler.getFunction and transform.Optimize.
	var abiNames []string
	if roots := mod.NamedGlobal("tinygo.indirect-abi"); !roots.IsNil() {
		builder := mod.Context().NewBuilder()
		initializer := roots.Initializer()
		for i := 0; i < initializer.Type().ArrayLength(); i++ {
			value := builder.CreateExtractValue(initializer, i, "")
			for !value.IsAConstantExpr().IsNil() && value.OperandsCount() == 1 {
				value = value.Operand(0)
			}
			abiNames = append(abiNames, value.Name())
		}
		builder.Dispose()
		llvmutil.RemoveGlobalReferences(mod, "llvm.used", "tinygo.indirect-abi")
	}

	options := llvm.NewPassBuilderOptions()
	defer options.Dispose()
	err := mod.RunPasses("globaldce", llvm.TargetMachine{}, options)
	var abiRoots []llvm.Value
	for _, name := range abiNames {
		if fn := mod.NamedFunction(name); !fn.IsNil() {
			abiRoots = append(abiRoots, fn)
		}
	}
	if len(abiRoots) != 0 {
		llvmutil.AppendToGlobal(mod, "llvm.used", abiRoots...)
		llvmutil.AppendToGlobal(mod, "tinygo.indirect-abi", abiRoots...)
	}
	if err != nil {
		return []error{fmt.Errorf("could not remove dead code before //go:noruntime check: %w", err)}
	}

	var errs []error
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if !fn.GetStringAttributeAtIndex(-1, "tinygo-noruntime").IsNil() {
			errs = append(errs, errorAt(fn, "//go:noruntime function "+fn.Name()+" remains reachable at runtime"))
		}
	}
	return errs
}

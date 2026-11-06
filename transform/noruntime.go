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
		if isNoRuntime(fn) {
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
		if !isNoRuntime(fn) {
			continue
		}
		errs = append(errs, errorAt(fn, "//go:noruntime function "+fn.Name()+" remains reachable at runtime"))
		if use := runtimeUse(fn, map[llvm.Value]bool{}); !use.IsNil() {
			caller := use.InstructionParent().Parent()
			if pos := getPosition(use); !pos.IsValid() {
				use = caller
			}
			errs = append(errs, errorAt(use, "//go:noruntime function "+fn.Name()+" is used at runtime by "+caller.Name()))
		}
	}
	return errs
}

func isNoRuntime(fn llvm.Value) bool {
	return !fn.GetStringAttributeAtIndex(-1, "tinygo-noruntime").IsNil()
}

// runtimeUse returns the first instruction outside a //go:noruntime function
// that uses value, directly or through constants and globals.
func runtimeUse(value llvm.Value, visited map[llvm.Value]bool) llvm.Value {
	visited[value] = true
	for use := value.FirstUse(); !use.IsNil(); use = use.NextUse() {
		user := use.User()
		if visited[user] {
			continue
		}
		if !user.IsAInstruction().IsNil() {
			if !isNoRuntime(user.InstructionParent().Parent()) {
				return user
			}
			continue
		}
		if found := runtimeUse(user, visited); !found.IsNil() {
			return found
		}
	}
	return llvm.Value{}
}

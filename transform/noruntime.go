package transform

import (
	"fmt"

	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

// CheckNoRuntime rejects marked functions that survive dead-code removal.
// Call after interp and before any inlining (builder.optimizeProgram).
func CheckNoRuntime(mod llvm.Module) []error {
	hasMarked := false
	for fn := range llvmutil.Functions(mod) {
		if isNoRuntime(fn) {
			hasMarked = true
			break
		}
	}
	if !hasMarked {
		return nil
	}

	restoreRoots := suspendIndirectABIRoots(mod)
	options := llvm.NewPassBuilderOptions()
	defer options.Dispose()
	err := mod.RunPasses("globaldce", llvm.TargetMachine{}, options)
	restoreRoots()
	if err != nil {
		return []error{fmt.Errorf("could not remove dead code before //go:noruntime check: %w", err)}
	}

	var errs []error
	for fn := range llvmutil.Functions(mod) {
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
	for user := range llvmutil.Uses(value) {
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

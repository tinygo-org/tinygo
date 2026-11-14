package transform

import (
	"go/scanner"
	"strings"

	"tinygo.org/x/go-llvm"
)

// The compiler calls a declaration with this suffix in place of the runtime
// function in each check of a //go:nopanic function.
const noPanicSuffix = "$nopanic"

// CheckNoPanic reports each check of a //go:nopanic function that the
// optimizer did not prove. Call it after all optimization passes.
func CheckNoPanic(mod llvm.Module) []error {
	var list scanner.ErrorList
	seen := map[string]bool{}
	for _, marker := range noPanicMarkers(mod) {
		for _, use := range getUses(marker) {
			err := errorAt(use, "//go:nopanic check is not proven")
			if !err.Pos.IsValid() {
				err.Msg += " in " + use.InstructionParent().Parent().Name()
			}
			// Inlining can copy one check into many functions.
			if key := err.Error(); !seen[key] {
				seen[key] = true
				list = append(list, &err)
			}
		}
		if marker.FirstUse().IsNil() {
			marker.EraseFromParentAsFunction()
		}
	}
	list.Sort()
	var errs []error
	for _, err := range list {
		errs = append(errs, *err)
	}
	return errs
}

// LowerNoPanic replaces each //go:nopanic check marker with the runtime
// function it stands for. Call it when the optimizer does not run.
func LowerNoPanic(mod llvm.Module) {
	for _, marker := range noPanicMarkers(mod) {
		fn := mod.NamedFunction(strings.TrimSuffix(marker.Name(), noPanicSuffix))
		marker.ReplaceAllUsesWith(fn)
		marker.EraseFromParentAsFunction()
	}
}

func noPanicMarkers(mod llvm.Module) []llvm.Value {
	var markers []llvm.Value
	for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
		if fn.IsDeclaration() && strings.HasSuffix(fn.Name(), noPanicSuffix) {
			markers = append(markers, fn)
		}
	}
	return markers
}

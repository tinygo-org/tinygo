package transform

import (
	"fmt"
	"go/token"
	"regexp"
	"sort"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"tinygo.org/x/go-llvm"
)

// Runtime functions called when a bounds check fails.
// The kind uses the names of the Go compiler's check_bce output.
var boundsPanics = []struct{ name, kind string }{
	{"runtime.lookupPanic", "IsInBounds"},
	{"runtime.slicePanic", "IsSliceInBounds"},
	{"runtime.sliceToArrayPointerPanic", "IsSliceInBounds"},
	{"runtime.unsafeSlicePanic", "IsSliceInBounds"},
}

// protectBoundsPanics stops the panic functions from being inlined so that
// remaining checks can be found after optimization.
func protectBoundsPanics(mod llvm.Module) {
	kind := llvm.AttributeKindID("noinline")
	for _, bp := range boundsPanics {
		fn := mod.NamedFunction(bp.name)
		if fn.IsNil() || fn.IsDeclaration() {
			continue
		}
		fn.AddFunctionAttr(mod.Context().CreateEnumAttribute(kind, 0))
	}
}

// PrintBoundsChecks reports the bounds checks left in the module. The regexp
// is matched against the function name and the source file of the check.
func PrintBoundsChecks(mod llvm.Module, re *regexp.Regexp, logger func(token.Position, string)) {
	type report struct {
		pos  token.Position
		kind string
	}
	var reports []report
	seen := map[report]struct{}{}
	for _, bp := range boundsPanics {
		fn := mod.NamedFunction(bp.name)
		for call := range llvmutil.Uses(fn) {
			if call.IsACallInst().IsNil() || call.CalledValue() != fn {
				continue
			}
			block := call.InstructionParent()
			fnName := block.Parent().Name()
			// Find the branches that lead to the throw block.
			for use := range llvmutil.Uses(block.AsValue()) {
				if use.IsABranchInst().IsNil() {
					continue
				}
				pos := getPosition(use)
				if !re.MatchString(fnName) && !re.MatchString(pos.Filename) {
					continue
				}
				r := report{pos, bp.kind}
				if _, ok := seen[r]; ok {
					continue
				}
				seen[r] = struct{}{}
				reports = append(reports, r)
			}
		}
	}
	sort.SliceStable(reports, func(i, j int) bool {
		a, b := reports[i].pos, reports[j].pos
		if a.Filename != b.Filename {
			return a.Filename < b.Filename
		}
		if a.Line != b.Line {
			return a.Line < b.Line
		}
		return a.Column < b.Column
	})
	for _, r := range reports {
		logger(r.pos, r.kind)
	}
}

// FormatBoundsCheck renders a bounds check like the Go compiler does.
func FormatBoundsCheck(pos token.Position, kind string) string {
	return fmt.Sprintf("%s: Found %s", pos.String(), kind)
}

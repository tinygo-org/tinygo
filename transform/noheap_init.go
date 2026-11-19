package transform

import (
	"strings"

	"github.com/tinygo-org/tinygo/internal/llvmutil"
	"tinygo.org/x/go-llvm"
)

// CheckNoHeapInit verifies that runtime.alloc, runtime.alloc_noheap, and indirect
// calls are transitively unreachable from runtime.hwInit0.
func CheckNoHeapInit(mod llvm.Module) []error {
	root := mod.NamedFunction("runtime.hwInit0")
	if root.IsNil() || root.IsDeclaration() {
		return nil
	}

	visited := map[llvm.Value]bool{root: true}
	queue := []llvm.Value{root}
	errs := []error{}

	for len(queue) > 0 {
		fn := queue[0]
		queue = queue[1:]
		for bb := range llvmutil.BasicBlocks(fn) {
			for inst := range llvmutil.Instructions(bb) {
				if inst.IsACallInst().IsNil() && inst.IsAInvokeInst().IsNil() {
					continue
				}
				callee := inst.CalledValue()
				if !callee.IsAInlineAsm().IsNil() {
					continue
				}
				if callee.IsAFunction().IsNil() {
					errs = append(errs, errorAt(inst, "indirect call is not allowed in pre-heap initialization (in "+fn.Name()+")"))
					continue
				}
				name := callee.Name()
				if strings.HasPrefix(name, "llvm.") {
					continue
				}
				if name == "runtime.alloc" || name == "runtime.alloc_noheap" {
					errs = append(errs, errorAt(inst, "heap allocation is not allowed in pre-heap initialization (in "+fn.Name()+")"))
					continue
				}
				if !callee.IsDeclaration() && !visited[callee] {
					visited[callee] = true
					queue = append(queue, callee)
				}
			}
		}
	}

	return errs
}

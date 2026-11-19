package transform

import (
	"strings"

	"tinygo.org/x/go-llvm"
)

// CheckNoHeapInit verifies that runtime.alloc* and indirect calls are
// transitively unreachable from runtime.hwInit0.
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
		for bb := fn.FirstBasicBlock(); !bb.IsNil(); bb = llvm.NextBasicBlock(bb) {
			for inst := bb.FirstInstruction(); !inst.IsNil(); inst = llvm.NextInstruction(inst) {
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
				if strings.HasPrefix(name, "runtime.alloc") {
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

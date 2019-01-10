package ir

import (
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// This file implements several optimization passes (analysis + transform) to
// optimize code in SSA form before it is compiled to LLVM IR. It is based on
// the IR defined in ir.go.

// Make a readable version of a method signature (including the function name,
// excluding the receiver name). This string is used internally to match
// interfaces and to call the correct method on an interface. Examples:
//
//     String() string
//     Read([]byte) (int, error)
func MethodSignature(method *types.Func) string {
	return method.Name() + signature(method.Type().(*types.Signature))
}

// Make a readable version of a function (pointer) signature.
// Examples:
//
//     () string
//     (string, int) (int, error)
func signature(sig *types.Signature) string {
	s := ""
	if sig.Params().Len() == 0 {
		s += "()"
	} else {
		s += "("
		for i := 0; i < sig.Params().Len(); i++ {
			if i > 0 {
				s += ", "
			}
			s += sig.Params().At(i).Type().String()
		}
		s += ")"
	}
	if sig.Results().Len() == 0 {
		// keep as-is
	} else if sig.Results().Len() == 1 {
		s += " " + sig.Results().At(0).Type().String()
	} else {
		s += " ("
		for i := 0; i < sig.Results().Len(); i++ {
			if i > 0 {
				s += ", "
			}
			s += sig.Results().At(i).Type().String()
		}
		s += ")"
	}
	return s
}

// Simple pass that removes dead code. This pass makes later analysis passes
// more useful.
func (p *Program) SimpleDCE() {
	// Unmark all functions.
	for _, f := range p.Functions {
		f.flag = false
	}

	// Initial set of live functions. Include main.main, *.init and runtime.*
	// functions.
	main := p.mainPkg.Members["main"].(*ssa.Function)
	runtimePkg := p.Program.ImportedPackage("runtime")
	p.GetFunction(main).flag = true
	worklist := []*ssa.Function{main}
	for _, f := range p.Functions {
		if f.exported || f.Synthetic == "package initializer" || f.Pkg == runtimePkg {
			if f.flag || isCGoInternal(f.Name()) {
				continue
			}
			f.flag = true
			worklist = append(worklist, f.Function)
		}
	}

	// Mark all called functions recursively.
	for len(worklist) != 0 {
		f := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		for _, block := range f.Blocks {
			for _, instr := range block.Instrs {
				if instr, ok := instr.(*ssa.MakeInterface); ok {
					for _, sel := range getAllMethods(p.Program, instr.X.Type()) {
						fn := p.Program.MethodValue(sel)
						callee := p.GetFunction(fn)
						if callee == nil {
							// TODO: why is this necessary?
							p.addFunction(fn)
							callee = p.GetFunction(fn)
						}
						if !callee.flag {
							callee.flag = true
							worklist = append(worklist, callee.Function)
						}
					}
				}
				for _, operand := range instr.Operands(nil) {
					if operand == nil || *operand == nil || isCGoInternal((*operand).Name()) {
						continue
					}
					switch operand := (*operand).(type) {
					case *ssa.Function:
						f := p.GetFunction(operand)
						if f == nil {
							// FIXME HACK: this function should have been
							// discovered already. It is not for bound methods.
							p.addFunction(operand)
							f = p.GetFunction(operand)
						}
						if !f.flag {
							f.flag = true
							worklist = append(worklist, operand)
						}
					}
				}
			}
		}
	}

	// Remove unmarked functions.
	livefunctions := []*Function{}
	for _, f := range p.Functions {
		if f.flag {
			livefunctions = append(livefunctions, f)
		} else {
			delete(p.functionMap, f.Function)
		}
	}
	p.Functions = livefunctions
}

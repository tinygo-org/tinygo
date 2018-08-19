package main

import (
	"go/types"
	"golang.org/x/tools/go/ssa"
)

// Make a readable version of the method signature (including the function name,
// excluding the receiver name). This string is used internally to match
// interfaces and to call the correct method on an interface. Examples:
//
//     String() string
//     Read([]byte) (int, error)
func MethodName(method *types.Func) string {
	sig := method.Type().(*types.Signature)
	name := method.Name()
	if sig.Params().Len() == 0 {
		name += "()"
	} else {
		name += "("
		for i := 0; i < sig.Params().Len(); i++ {
			if i > 0 {
				name += ", "
			}
			name += sig.Params().At(i).Type().String()
		}
		name += ")"
	}
	if sig.Results().Len() == 0 {
		// keep as-is
	} else if sig.Results().Len() == 1 {
		name += " " + sig.Results().At(0).Type().String()
	} else {
		name += " ("
		for i := 0; i < sig.Results().Len(); i++ {
			if i > 0 {
				name += ", "
			}
			name += sig.Results().At(i).Type().String()
		}
		name += ")"
	}
	return name
}

// Fill in parents of all functions.
//
// All packages need to be added before this pass can run, or it will produce
// incorrect results.
func (p *Program) AnalyseCallgraph() {
	for _, f := range p.Functions {
		// Clear, if AnalyseCallgraph has been called before.
		f.children = nil
		f.parents = nil

		for _, block := range f.fn.Blocks {
			for _, instr := range block.Instrs {
				switch instr := instr.(type) {
				case *ssa.Call:
					if instr.Common().IsInvoke() {
						continue
					}
					switch call := instr.Call.Value.(type) {
					case *ssa.Builtin:
						// ignore
					case *ssa.Function:
						if isCGoInternal(call.Name()) {
							continue
						}
						child := p.GetFunction(call)
						if child.CName() != "" {
							continue // assume non-blocking
						}
						if child.LinkName(false) == "runtime.Sleep" {
							f.blocking = true
						}
						f.children = append(f.children, child)
					}
				}
			}
		}
	}
	for _, f := range p.Functions {
		for _, child := range f.children {
			child.parents = append(child.parents, f)
		}
	}
}

// Find all types that are put in an interface.
func (p *Program) AnalyseInterfaceConversions() {
	// Clear, if AnalyseTypes has been called before.
	p.typesWithMethods = make(map[string]*InterfaceType)
	p.typesWithoutMethods = make(map[string]int)

	for _, f := range p.Functions {
		for _, block := range f.fn.Blocks {
			for _, instr := range block.Instrs {
				switch instr := instr.(type) {
				case *ssa.MakeInterface:
					methods := getAllMethods(f.fn.Prog, instr.X.Type())
					name := instr.X.Type().String()
					if _, ok := p.typesWithMethods[name]; !ok && len(methods) > 0 {
						t := &InterfaceType{
							t:       instr.X.Type(),
							Num:     len(p.typesWithMethods),
							Methods: make(map[string]*types.Selection),
						}
						for _, sel := range methods {
							name := MethodName(sel.Obj().(*types.Func))
							t.Methods[name] = sel
						}
						p.typesWithMethods[instr.X.Type().String()] = t
					} else if _, ok := p.typesWithoutMethods[name]; !ok && len(methods) == 0 {
						p.typesWithoutMethods[name] = len(p.typesWithoutMethods)
					}
				}
			}
		}
	}
}

// Analyse which functions are recursively blocking.
//
// Depends on AnalyseCallgraph.
func (p *Program) AnalyseBlockingRecursive() {
	worklist := make([]*Function, 0)

	// Fill worklist with directly blocking functions.
	for _, f := range p.Functions {
		if f.blocking {
			worklist = append(worklist, f)
		}
	}

	// Keep reducing this worklist by marking a function as recursively blocking
	// from the worklist and pushing all its parents that are non-blocking.
	// This is somewhat similar to a worklist in a mark-sweep garbage collector.
	// The work items are then grey objects.
	for len(worklist) != 0 {
		// Pick the topmost.
		f := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		for _, parent := range f.parents {
			if !parent.blocking {
				parent.blocking = true
				worklist = append(worklist, parent)
			}
		}
	}
}

// Check whether we need a scheduler. A scheduler is only necessary when there
// are go calls that start blocking functions (if they're not blocking, the go
// function can be turned into a regular function call).
//
// Depends on AnalyseBlockingRecursive.
func (p *Program) AnalyseGoCalls() {
	p.goCalls = nil
	for _, f := range p.Functions {
		for _, block := range f.fn.Blocks {
			for _, instr := range block.Instrs {
				switch instr := instr.(type) {
				case *ssa.Go:
					p.goCalls = append(p.goCalls, instr)
				}
			}
		}
	}
	for _, instr := range p.goCalls {
		switch instr := instr.Call.Value.(type) {
		case *ssa.Builtin:
		case *ssa.Function:
			if p.functionMap[instr].blocking {
				p.needsScheduler = true
			}
		default:
			panic("unknown go call function type")
		}
	}
}

// Simple pass that removes dead code. This pass makes later analysis passes
// more useful.
func (p *Program) SimpleDCE() {
	// Unmark all functions and globals.
	for _, f := range p.Functions {
		f.flag = false
	}
	for _, f := range p.Globals {
		f.flag = false
	}

	// Initial set of live functions. Include main.main, *.init and runtime.*
	// functions.
	main := p.mainPkg.Members["main"].(*ssa.Function)
	runtimePkg := p.program.ImportedPackage("runtime")
	p.GetFunction(main).flag = true
	worklist := []*ssa.Function{main}
	for _, f := range p.Functions {
		if f.fn.Synthetic == "package initializer" || f.fn.Pkg == runtimePkg {
			if f.flag || isCGoInternal(f.fn.Name()) {
				continue
			}
			f.flag = true
			worklist = append(worklist, f.fn)
		}
	}

	// Mark all called functions recursively.
	for len(worklist) != 0 {
		f := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		for _, block := range f.Blocks {
			for _, instr := range block.Instrs {
				if instr, ok := instr.(*ssa.MakeInterface); ok {
					for _, sel := range getAllMethods(p.program, instr.X.Type()) {
						callee := p.GetFunction(p.program.MethodValue(sel))
						if !callee.flag {
							callee.flag = true
							worklist = append(worklist, callee.fn)
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
						if !f.flag {
							f.flag = true
							worklist = append(worklist, operand)
						}
					case *ssa.Global:
						// TODO: globals that reference other globals
						global := p.GetGlobal(operand)
						global.flag = true
					}
				}
			}
		}
	}

	// Remove unmarked functions.
	livefunctions := []*Function{p.GetFunction(main)}
	for _, f := range p.Functions {
		if f.flag {
			livefunctions = append(livefunctions, f)
		} else {
			delete(p.functionMap, f.fn)
		}
	}
	p.Functions = livefunctions

	// Remove unmarked globals.
	liveglobals := []*Global{}
	for _, g := range p.Globals {
		if g.flag {
			liveglobals = append(liveglobals, g)
		} else {
			delete(p.globalMap, g.g)
		}
	}
	p.Globals = liveglobals
}

// Whether this function needs a scheduler.
//
// Depends on AnalyseGoCalls.
func (p *Program) NeedsScheduler() bool {
	return p.needsScheduler
}

// Whether this function blocks. Builtins are also accepted for convenience.
// They will always be non-blocking.
//
// Depends on AnalyseBlockingRecursive.
func (p *Program) IsBlocking(f *Function) bool {
	if !p.needsScheduler {
		return false
	}
	return f.blocking
}

// Return the type number and whether this type is actually used. Used in
// interface conversions (type is always used) and type asserts (type may not be
// used, meaning assert is always false in this program).
//
// May only be used after all packages have been added to the analyser.
func (p *Program) TypeNum(typ types.Type) (int, bool) {
	if n, ok := p.typesWithoutMethods[typ.String()]; ok {
		return n, true
	} else if meta, ok := p.typesWithMethods[typ.String()]; ok {
		return len(p.typesWithoutMethods) + meta.Num, true
	} else {
		return -1, false // type is never put in an interface
	}
}

// MethodNum returns the numeric ID of this method, to be used in method lookups
// on interfaces for example.
func (p *Program) MethodNum(method *types.Func) int {
	name := MethodName(method)
	if _, ok := p.methodSignatureNames[name]; !ok {
		p.methodSignatureNames[name] = len(p.methodSignatureNames)
	}
	return p.methodSignatureNames[MethodName(method)]
}

// The start index of the first dynamic type that has methods.
// Types without methods always have a lower ID and types with methods have this
// or a higher ID.
//
// May only be used after all packages have been added to the analyser.
func (p *Program) FirstDynamicType() int {
	return len(p.typesWithoutMethods)
}

// Return all types with methods, sorted by type ID.
func (p *Program) AllDynamicTypes() []*InterfaceType {
	l := make([]*InterfaceType, len(p.typesWithMethods))
	for _, m := range p.typesWithMethods {
		l[m.Num] = m
	}
	return l
}

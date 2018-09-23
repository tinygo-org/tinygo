package ir

import (
	"go/types"
	"sort"
	"strings"

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
	return method.Name() + Signature(method.Type().(*types.Signature))
}

// Make a readable version of a function (pointer) signature. This string is
// used internally to match signatures (like in AnalyseFunctionPointers).
// Examples:
//
//     () string
//     (string, int) (int, error)
func Signature(sig *types.Signature) string {
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

// Convert an interface type to a string of all method strings, separated by
// "; ". For example: "Read([]byte) (int, error); Close() error"
func InterfaceKey(itf *types.Interface) string {
	methodStrings := []string{}
	for i := 0; i < itf.NumMethods(); i++ {
		method := itf.Method(i)
		methodStrings = append(methodStrings, MethodSignature(method))
	}
	sort.Strings(methodStrings)
	return strings.Join(methodStrings, ";")
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

		for _, block := range f.Blocks {
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
						if child.RelString(nil) == "time.Sleep" {
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
	p.typesWithoutMethods = map[string]int{"nil": 0}
	p.typesWithMethods = map[string]*TypeWithMethods{}

	for _, f := range p.Functions {
		for _, block := range f.Blocks {
			for _, instr := range block.Instrs {
				switch instr := instr.(type) {
				case *ssa.MakeInterface:
					methods := getAllMethods(f.Prog, instr.X.Type())
					name := instr.X.Type().String()
					if _, ok := p.typesWithMethods[name]; !ok && len(methods) > 0 {
						t := &TypeWithMethods{
							t:       instr.X.Type(),
							Num:     len(p.typesWithMethods),
							Methods: make(map[string]*types.Selection),
						}
						for _, sel := range methods {
							name := MethodSignature(sel.Obj().(*types.Func))
							t.Methods[name] = sel
						}
						p.typesWithMethods[name] = t
					} else if _, ok := p.typesWithoutMethods[name]; !ok && len(methods) == 0 {
						p.typesWithoutMethods[name] = len(p.typesWithoutMethods)
					}
				}
			}
		}
	}
}

// Analyse which function pointer signatures need a context parameter.
// This makes calling function pointers more efficient.
func (p *Program) AnalyseFunctionPointers() {
	// Clear, if AnalyseFunctionPointers has been called before.
	p.fpWithContext = map[string]struct{}{}

	for _, f := range p.Functions {
		for _, block := range f.Blocks {
			for _, instr := range block.Instrs {
				switch instr := instr.(type) {
				case ssa.CallInstruction:
					for _, arg := range instr.Common().Args {
						switch arg := arg.(type) {
						case *ssa.Function:
							f := p.GetFunction(arg)
							f.addressTaken = true
						}
					}
				case *ssa.DebugRef:
				default:
					// For anything that isn't a call...
					for _, operand := range instr.Operands(nil) {
						if operand == nil || *operand == nil || isCGoInternal((*operand).Name()) {
							continue
						}
						switch operand := (*operand).(type) {
						case *ssa.Function:
							f := p.GetFunction(operand)
							f.addressTaken = true
						}
					}
				}
				switch instr := instr.(type) {
				case *ssa.MakeClosure:
					fn := instr.Fn.(*ssa.Function)
					sig := Signature(fn.Signature)
					p.fpWithContext[sig] = struct{}{}
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
		for _, block := range f.Blocks {
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

// InterfaceNum returns the numeric interface ID of this type, for use in type
// asserts.
func (p *Program) InterfaceNum(itfType *types.Interface) int {
	key := InterfaceKey(itfType)
	if itf, ok := p.interfaces[key]; !ok {
		num := len(p.interfaces)
		p.interfaces[key] = &Interface{Num: num, Type: itfType}
		return num
	} else {
		return itf.Num
	}
}

// MethodNum returns the numeric ID of this method, to be used in method lookups
// on interfaces for example.
func (p *Program) MethodNum(method *types.Func) int {
	name := MethodSignature(method)
	if _, ok := p.methodSignatureNames[name]; !ok {
		p.methodSignatureNames[name] = len(p.methodSignatureNames)
	}
	return p.methodSignatureNames[MethodSignature(method)]
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
func (p *Program) AllDynamicTypes() []*TypeWithMethods {
	l := make([]*TypeWithMethods, len(p.typesWithMethods))
	for _, m := range p.typesWithMethods {
		l[m.Num] = m
	}
	return l
}

// Return all interface types, sorted by interface ID.
func (p *Program) AllInterfaces() []*Interface {
	l := make([]*Interface, len(p.interfaces))
	for _, itf := range p.interfaces {
		l[itf.Num] = itf
	}
	return l
}

func (p *Program) FunctionNeedsContext(f *Function) bool {
	if !f.addressTaken {
		return false
	}
	return p.SignatureNeedsContext(f.Signature)
}

func (p *Program) SignatureNeedsContext(sig *types.Signature) bool {
	_, needsContext := p.fpWithContext[Signature(sig)]
	return needsContext
}

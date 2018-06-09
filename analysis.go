
package main

import (
	"go/types"

	"golang.org/x/tools/go/ssa"
)

// Analysis results over a whole program.
type Analysis struct {
	functions            map[*ssa.Function]*FuncMeta
	needsScheduler       bool
	goCalls              []*ssa.Go
	typesWithMethods     map[string]*TypeMeta
	typesWithoutMethods  map[string]int
	methodSignatureNames map[string]int
}

// Some analysis results of a single function.
type FuncMeta struct {
	f                 *ssa.Function
	blocking          bool
	parents           []*ssa.Function // calculated by AnalyseCallgraph
	children          []*ssa.Function
}

type TypeMeta struct {
	t       types.Type
	Num     int
	Methods map[string]*types.Selection
}

// Return a new Analysis object.
func NewAnalysis() *Analysis {
	return &Analysis{
		functions:            make(map[*ssa.Function]*FuncMeta),
		typesWithMethods:     make(map[string]*TypeMeta),
		typesWithoutMethods:  make(map[string]int),
		methodSignatureNames: make(map[string]int),
	}
}

// Add a given package to the analyzer, to be analyzed later.
func (a *Analysis) AddPackage(pkg *ssa.Package) {
	for _, member := range pkg.Members {
		switch member := member.(type) {
		case *ssa.Function:
			if isCGoInternal(member.Name()) || getCName(member.Name()) != "" {
				continue
			}
			a.addFunction(member)
		case *ssa.Type:
			methods := getAllMethods(pkg.Prog, member.Type())
			if types.IsInterface(member.Type()) {
				for _, method := range methods {
					a.MethodName(method.Obj().(*types.Func))
				}
			} else { // named type
				for _, method := range methods {
					a.addFunction(pkg.Prog.MethodValue(method))
				}
			}
		}
	}
}

// Analyze the given function quickly without any recursion, and add it to the
// list of functions in the analyzer.
func (a *Analysis) addFunction(f *ssa.Function) {
	fm := &FuncMeta{}
	for _, block := range f.Blocks {
		for _, instr := range block.Instrs {
			switch instr := instr.(type) {
			case *ssa.Call:
				if instr.Common().IsInvoke() {
					name := a.MethodName(instr.Common().Method)
					a.methodSignatureNames[name] = len(a.methodSignatureNames)
				} else {
					switch call := instr.Call.Value.(type) {
					case *ssa.Builtin:
						// ignore
					case *ssa.Function:
						if isCGoInternal(call.Name()) || getCName(call.Name()) != "" {
							continue
						}
						name := getFunctionName(call, false)
						if name == "runtime.Sleep" {
							fm.blocking = true
						}
						fm.children = append(fm.children, call)
					}
				}
			case *ssa.MakeInterface:
				methods := getAllMethods(f.Prog, instr.X.Type())
				if _, ok := a.typesWithMethods[instr.X.Type().String()]; !ok && len(methods) > 0 {
					meta := &TypeMeta{
						t:       instr.X.Type(),
						Num:     len(a.typesWithMethods),
						Methods: make(map[string]*types.Selection),
					}
					for _, sel := range methods {
						name := a.MethodName(sel.Obj().(*types.Func))
						meta.Methods[name] = sel
					}
					a.typesWithMethods[instr.X.Type().String()] = meta
				} else if _, ok := a.typesWithoutMethods[instr.X.Type().String()]; !ok && len(methods) == 0 {
					a.typesWithoutMethods[instr.X.Type().String()] = len(a.typesWithoutMethods)
				}
			case *ssa.Go:
				a.goCalls = append(a.goCalls, instr)
			}
		}
	}
	a.functions[f] = fm

	for _, child := range f.AnonFuncs {
		a.addFunction(child)
	}
}

// Make a readable version of the method signature (including the function name,
// excluding the receiver name). This string is used internally to match
// interfaces and to call the correct method on an interface. Examples:
//
//     String() string
//     Read([]byte) (int, error)
func (a *Analysis) MethodName(method *types.Func) string {
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
func (a *Analysis) AnalyseCallgraph() {
	for f, fm := range a.functions {
		for _, child := range fm.children {
			childRes, ok := a.functions[child]
			if !ok {
				println("child not found: " + child.Pkg.Pkg.Path() + "." + child.Name() + ", function: " + f.Name())
				continue
			}
			childRes.parents = append(childRes.parents, f)
		}
	}
}

// Analyse which functions are recursively blocking.
//
// Depends on AnalyseCallgraph.
func (a *Analysis) AnalyseBlockingRecursive() {
	worklist := make([]*FuncMeta, 0)

	// Fill worklist with directly blocking functions.
	for _, fm := range a.functions {
		if fm.blocking {
			worklist = append(worklist, fm)
		}
	}

	// Keep reducing this worklist by marking a function as recursively blocking
	// from the worklist and pushing all its parents that are non-blocking.
	// This is somewhat similar to a worklist in a mark-sweep garbage collector.
	// The work items are then grey objects.
	for len(worklist) != 0 {
		// Pick the topmost.
		fm := worklist[len(worklist)-1]
		worklist = worklist[:len(worklist)-1]
		for _, parent := range fm.parents {
			parentfm := a.functions[parent]
			if !parentfm.blocking {
				parentfm.blocking = true
				worklist = append(worklist, parentfm)
			}
		}
	}
}

// Check whether we need a scheduler. A scheduler is only necessary when there
// are go calls that start blocking functions (if they're not blocking, the go
// function can be turned into a regular function call).
//
// Depends on AnalyseBlockingRecursive.
func (a *Analysis) AnalyseGoCalls() {
	for _, instr := range a.goCalls {
		if a.isBlocking(instr.Call.Value) {
			a.needsScheduler = true
		}
	}
}

// Whether this function needs a scheduler.
//
// Depends on AnalyseGoCalls.
func (a *Analysis) NeedsScheduler() bool {
	return a.needsScheduler
}

// Whether this function blocks. Builtins are also accepted for convenience.
// They will always be non-blocking.
//
// Depends on AnalyseBlockingRecursive.
func (a *Analysis) IsBlocking(f ssa.Value) bool {
	if !a.needsScheduler {
		return false
	}
	return a.isBlocking(f)
}

func (a *Analysis) isBlocking(f ssa.Value) bool {
	switch f := f.(type) {
	case *ssa.Builtin:
		return false
	case *ssa.Function:
		return a.functions[f].blocking
	default:
		panic("Analysis.IsBlocking on unknown type")
	}
}

// Return the type number and whether this type is actually used. Used in
// interface conversions (type is always used) and type asserts (type may not be
// used, meaning assert is always false in this program).
//
// May only be used after all packages have been added to the analyser.
func (a *Analysis) TypeNum(typ types.Type) (int, bool) {
	if n, ok := a.typesWithoutMethods[typ.String()]; ok {
		return n, true
	} else if meta, ok := a.typesWithMethods[typ.String()]; ok {
		return len(a.typesWithoutMethods) + meta.Num, true
	} else {
		return -1, false // type is never put in an interface
	}
}

// MethodNum returns the numeric ID of this method, to be used in method lookups
// on interfaces for example.
func (a *Analysis) MethodNum(method *types.Func) int {
	if n, ok := a.methodSignatureNames[a.MethodName(method)]; ok {
		return n
	}
	return -1 // signal error
}

// The start index of the first dynamic type that has methods.
// Types without methods always have a lower ID and types with methods have this
// or a higher ID.
//
// May only be used after all packages have been added to the analyser.
func (a *Analysis) FirstDynamicType() int {
	return len(a.typesWithoutMethods)
}

// Return all types with methods, sorted by type ID.
func (a *Analysis) AllDynamicTypes() []*TypeMeta {
	l := make([]*TypeMeta, len(a.typesWithMethods))
	for _, m := range a.typesWithMethods {
		l[m.Num] = m
	}
	return l
}


package main

import (
	"golang.org/x/tools/go/ssa"
)

// Analysis results over a whole program.
type Analysis struct {
	functions      map[*ssa.Function]*FuncMeta
	needsScheduler bool
	goCalls        []*ssa.Go
}

// Some analysis results of a single function.
type FuncMeta struct {
	f                 *ssa.Function
	blocking          bool
	parents           []*ssa.Function // calculated by AnalyseCallgraph
	children          []*ssa.Function
}

// Return a new Analysis object.
func NewAnalysis() *Analysis {
	return &Analysis{
		functions: make(map[*ssa.Function]*FuncMeta),
	}
}

// Add a given package to the analyzer, to be analyzed later.
func (a *Analysis) AddPackage(pkg *ssa.Package) {
	for _, member := range pkg.Members {
		switch member := member.(type) {
		case *ssa.Function:
			a.addFunction(member)
		case *ssa.Type:
			ms := pkg.Prog.MethodSets.MethodSet(member.Type())
			for i := 0; i < ms.Len(); i++ {
				a.addFunction(pkg.Prog.MethodValue(ms.At(i)))
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
				switch call := instr.Call.Value.(type) {
				case *ssa.Function:
					name := getFunctionName(call, false)
					if name == "runtime.Sleep" {
						fm.blocking = true
					}
					fm.children = append(fm.children, call)
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

// Fill in parents of all functions.
//
// All packages need to be added before this pass can run, or it will produce
// incorrect results.
func (a *Analysis) AnalyseCallgraph() {
	for f, fm := range a.functions {
		for _, child := range fm.children {
			childRes, ok := a.functions[child]
			if !ok {
				print("child not found: " + child.Pkg.Pkg.Path() + "." + child.Name() + ", function: " + f.Name())
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

// Check whether we need a scheduler. This is only necessary when there are go
// calls that start blocking functions (if they're not blocking, the go function
// can be turned into a regular function call).
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

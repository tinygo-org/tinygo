package compiler

// This file implements a simple reachability analysis, to reduce compile time.
// This DCE pass used to be necessary for improving other passes but now it
// isn't necessary anymore.

import (
	"errors"
	"go/types"
	"sort"

	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/ssa"
)

type dceState struct {
	*compilerContext
	functions   []*dceFunction
	functionMap map[*ssa.Function]*dceFunction
}

type dceFunction struct {
	*ssa.Function
	functionInfo
	flag bool // used by dead code elimination
}

func (p *dceState) addFunction(ssaFn *ssa.Function) {
	if _, ok := p.functionMap[ssaFn]; ok {
		return
	}
	f := &dceFunction{Function: ssaFn}
	f.functionInfo = p.getFunctionInfo(ssaFn)
	p.functions = append(p.functions, f)
	p.functionMap[ssaFn] = f

	for _, anon := range ssaFn.AnonFuncs {
		p.addFunction(anon)
	}
}

// simpleDCE returns a list of alive functions in the program. Compiling only
// these functions makes the compiler faster.
//
// This functionality will likely be replaced in the future with build caching.
func (c *compilerContext) simpleDCE(lprogram *loader.Program) ([]*ssa.Function, error) {
	mainPkg := c.program.Package(lprogram.MainPkg().Pkg)
	if mainPkg == nil {
		panic("could not find main package")
	}
	p := &dceState{
		compilerContext: c,
		functionMap:     make(map[*ssa.Function]*dceFunction),
	}

	for _, pkg := range lprogram.Sorted() {
		pkg := c.program.Package(pkg.Pkg)
		memberNames := make([]string, 0)
		for name := range pkg.Members {
			memberNames = append(memberNames, name)
		}
		sort.Strings(memberNames)

		for _, name := range memberNames {
			member := pkg.Members[name]
			switch member := member.(type) {
			case *ssa.Function:
				p.addFunction(member)
			case *ssa.Type:
				methods := getAllMethods(pkg.Prog, member.Type())
				if !types.IsInterface(member.Type()) {
					// named type
					for _, method := range methods {
						p.addFunction(pkg.Prog.MethodValue(method))
					}
				}
			case *ssa.Global:
				// Ignore. Globals are not handled here.
			case *ssa.NamedConst:
				// Ignore: these are already resolved.
			default:
				panic("unknown member type: " + member.String())
			}
		}
	}

	// Initial set of live functions. Include main.main, *.init and runtime.*
	// functions.
	main, ok := mainPkg.Members["main"].(*ssa.Function)
	if !ok {
		if mainPkg.Members["main"] == nil {
			return nil, errors.New("function main is undeclared in the main package")
		} else {
			return nil, errors.New("cannot declare main - must be func")
		}
	}
	runtimePkg := c.program.ImportedPackage("runtime")
	mathPkg := c.program.ImportedPackage("math")
	taskPkg := c.program.ImportedPackage("internal/task")
	p.functionMap[main].flag = true
	worklist := []*ssa.Function{main}
	for _, f := range p.functions {
		if f.exported || f.Synthetic == "package initializer" || f.Pkg == runtimePkg || f.Pkg == taskPkg || (f.Pkg == mathPkg && f.Pkg != nil) {
			if f.flag {
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
					for _, sel := range getAllMethods(c.program, instr.X.Type()) {
						fn := c.program.MethodValue(sel)
						callee := p.functionMap[fn]
						if callee == nil {
							// TODO: why is this necessary?
							p.addFunction(fn)
							callee = p.functionMap[fn]
						}
						if !callee.flag {
							callee.flag = true
							worklist = append(worklist, callee.Function)
						}
					}
				}
				for _, operand := range instr.Operands(nil) {
					if operand == nil || *operand == nil {
						continue
					}
					switch operand := (*operand).(type) {
					case *ssa.Function:
						f := p.functionMap[operand]
						if f == nil {
							// FIXME HACK: this function should have been
							// discovered already. It is not for bound methods.
							p.addFunction(operand)
							f = p.functionMap[operand]
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

	// Return all live functions.
	liveFunctions := []*ssa.Function{}
	for _, f := range p.functions {
		if f.flag {
			liveFunctions = append(liveFunctions, f.Function)
		}
	}
	return liveFunctions, nil
}

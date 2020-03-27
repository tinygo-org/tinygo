package ir

import (
	"go/types"

	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/ssa"
)

// This file provides a wrapper around go/ssa values and adds extra
// functionality to them.

// View on all functions, types, and globals in a program, with analysis
// results.
type Program struct {
	Program       *ssa.Program
	LoaderProgram *loader.Program
	mainPkg       *ssa.Package
	mainPath      string
}

// Create and initialize a new *Program from a *ssa.Program.
func NewProgram(lprogram *loader.Program, mainPath string) *Program {
	program := lprogram.LoadSSA()
	program.Build()

	// Find the main package, which is a bit difficult when running a .go file
	// directly.
	mainPkg := program.ImportedPackage(mainPath)
	if mainPkg == nil {
		for _, pkgInfo := range program.AllPackages() {
			if pkgInfo.Pkg.Name() == "main" {
				if mainPkg != nil {
					panic("more than one main package found")
				}
				mainPkg = pkgInfo
			}
		}
	}
	if mainPkg == nil {
		panic("could not find main package")
	}

	return &Program{
		Program:       program,
		LoaderProgram: lprogram,
		mainPkg:       mainPkg,
		mainPath:      mainPath,
	}
}

// Packages returns a list of all packages, sorted by import order.
func (p *Program) Packages() []*ssa.Package {
	packageList := []*ssa.Package{}
	packageSet := map[string]struct{}{}
	worklist := []string{"runtime", p.mainPath}
	for len(worklist) != 0 {
		pkgPath := worklist[0]
		var pkg *ssa.Package
		if pkgPath == p.mainPath {
			pkg = p.mainPkg // necessary for compiling individual .go files
		} else {
			pkg = p.Program.ImportedPackage(pkgPath)
		}
		if pkg == nil {
			// Non-SSA package (e.g. cgo).
			packageSet[pkgPath] = struct{}{}
			worklist = worklist[1:]
			continue
		}
		if _, ok := packageSet[pkgPath]; ok {
			// Package already in the final package list.
			worklist = worklist[1:]
			continue
		}

		unsatisfiedImports := make([]string, 0)
		imports := pkg.Pkg.Imports()
		for _, pkg := range imports {
			if _, ok := packageSet[pkg.Path()]; ok {
				continue
			}
			unsatisfiedImports = append(unsatisfiedImports, pkg.Path())
		}
		if len(unsatisfiedImports) == 0 {
			// All dependencies of this package are satisfied, so add this
			// package to the list.
			packageList = append(packageList, pkg)
			packageSet[pkgPath] = struct{}{}
			worklist = worklist[1:]
		} else {
			// Prepend all dependencies to the worklist and reconsider this
			// package (by not removing it from the worklist). At that point, it
			// must be possible to add it to packageList.
			worklist = append(unsatisfiedImports, worklist...)
		}
	}

	return packageList
}

func (p *Program) MainPkg() *ssa.Package {
	return p.mainPkg
}

// MethodSignature creates a readable version of a method signature (including
// the function name, excluding the receiver name). This string is used
// internally to match interfaces and to call the correct method on an
// interface. Examples:
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

package loader

import (
	"golang.org/x/tools/go/ssa"
)

// LoadSSA constructs the SSA form of the loaded packages.
//
// The program must already be parsed and type-checked with the .Parse() method.
func (p *Program) LoadSSA() *ssa.Program {
	prog := ssa.NewProgram(p.fset, ssa.SanityCheckFunctions|ssa.BareInits|ssa.GlobalDebug)

	for _, pkg := range p.sorted {
		prog.CreatePackage(pkg.Pkg, pkg.Files, &pkg.info, true)
	}

	return prog
}

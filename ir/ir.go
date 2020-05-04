package ir

import (
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

// This file provides a wrapper around go/ssa values and adds extra
// functionality to them.

// View on all functions, types, and globals in a program, with analysis
// results.
type Program struct {
	Program       *ssa.Program
	LoaderProgram *loader.Program
	mainPkg       *ssa.Package
	Functions     []*Function
	functionMap   map[*ssa.Function]*Function
}

// Function or method.
type Function struct {
	*ssa.Function
	LLVMFn   llvm.Value
	module   string     // go:wasm-module
	linkName string     // go:linkname, go:export
	exported bool       // go:export
	nobounds bool       // go:nobounds
	flag     bool       // used by dead code elimination
	inline   InlineType // go:inline
}

// Interface type that is at some point used in a type assert (to check whether
// it implements another interface).
type Interface struct {
	Num  int
	Type *types.Interface
}

type InlineType int

// How much to inline.
const (
	// Default behavior. The compiler decides for itself whether any given
	// function will be inlined. Whether any function is inlined depends on the
	// optimization level.
	InlineDefault InlineType = iota

	// Inline hint, just like the C inline keyword (signalled using
	// //go:inline). The compiler will be more likely to inline this function,
	// but it is not a guarantee.
	InlineHint

	// Don't inline, just like the GCC noinline attribute. Signalled using
	// //go:noinline.
	InlineNone
)

// Create and initialize a new *Program from a *ssa.Program.
func NewProgram(lprogram *loader.Program) *Program {
	program := lprogram.LoadSSA()
	program.Build()

	// Find the main package, which is a bit difficult when running a .go file
	// directly.
	mainPkg := program.ImportedPackage(lprogram.MainPkg.PkgPath)
	if mainPkg == nil {
		panic("could not find main package")
	}

	// Make a list of packages in import order.
	packageList := []*ssa.Package{}
	packageSet := map[string]struct{}{}
	worklist := []string{"runtime", lprogram.MainPkg.PkgPath}
	for len(worklist) != 0 {
		pkgPath := worklist[0]
		pkg := program.ImportedPackage(pkgPath)
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

	p := &Program{
		Program:       program,
		LoaderProgram: lprogram,
		mainPkg:       mainPkg,
		functionMap:   make(map[*ssa.Function]*Function),
	}

	for _, pkg := range packageList {
		p.AddPackage(pkg)
	}

	return p
}

// Add a package to this Program. All packages need to be added first before any
// analysis is done for correct results.
func (p *Program) AddPackage(pkg *ssa.Package) {
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

func (p *Program) addFunction(ssaFn *ssa.Function) {
	if _, ok := p.functionMap[ssaFn]; ok {
		return
	}
	f := &Function{Function: ssaFn}
	f.parsePragmas()
	p.Functions = append(p.Functions, f)
	p.functionMap[ssaFn] = f

	for _, anon := range ssaFn.AnonFuncs {
		p.addFunction(anon)
	}
}

// Return true if this package imports "unsafe", false otherwise.
func hasUnsafeImport(pkg *types.Package) bool {
	for _, imp := range pkg.Imports() {
		if imp == types.Unsafe {
			return true
		}
	}
	return false
}

func (p *Program) GetFunction(ssaFn *ssa.Function) *Function {
	return p.functionMap[ssaFn]
}

func (p *Program) MainPkg() *ssa.Package {
	return p.mainPkg
}

// Parse compiler directives in the preceding comments.
func (f *Function) parsePragmas() {
	if f.Syntax() == nil {
		return
	}
	if decl, ok := f.Syntax().(*ast.FuncDecl); ok && decl.Doc != nil {
		for _, comment := range decl.Doc.List {
			text := comment.Text
			if strings.HasPrefix(text, "//export ") {
				// Rewrite '//export' to '//go:export' for compatibility with
				// gc.
				text = "//go:" + text[2:]
			}
			if !strings.HasPrefix(text, "//go:") {
				continue
			}
			parts := strings.Fields(text)
			switch parts[0] {
			case "//go:export":
				if len(parts) != 2 {
					continue
				}
				f.linkName = parts[1]
				f.exported = true
			case "//go:wasm-module":
				// Alternative comment for setting the import module.
				if len(parts) != 2 {
					continue
				}
				f.module = parts[1]
			case "//go:inline":
				f.inline = InlineHint
			case "//go:noinline":
				f.inline = InlineNone
			case "//go:linkname":
				if len(parts) != 3 || parts[1] != f.Name() {
					continue
				}
				// Only enable go:linkname when the package imports "unsafe".
				// This is a slightly looser requirement than what gc uses: gc
				// requires the file to import "unsafe", not the package as a
				// whole.
				if hasUnsafeImport(f.Pkg.Pkg) {
					f.linkName = parts[2]
				}
			case "//go:nobounds":
				// Skip bounds checking in this function. Useful for some
				// runtime functions.
				// This is somewhat dangerous and thus only imported in packages
				// that import unsafe.
				if hasUnsafeImport(f.Pkg.Pkg) {
					f.nobounds = true
				}
			}
		}
	}
}

func (f *Function) IsNoBounds() bool {
	return f.nobounds
}

// Return true iff this function is externally visible.
func (f *Function) IsExported() bool {
	return f.exported || f.CName() != ""
}

// Return the inline directive of this function.
func (f *Function) Inline() InlineType {
	return f.inline
}

// Return the module name if not the default.
func (f *Function) Module() string {
	return f.module
}

// Return the link name for this function.
func (f *Function) LinkName() string {
	if f.linkName != "" {
		return f.linkName
	}
	if f.Signature.Recv() != nil {
		// Method on a defined type (which may be a pointer).
		return f.RelString(nil)
	} else {
		// Bare function.
		if name := f.CName(); name != "" {
			// Name CGo functions directly.
			return name
		} else {
			return f.RelString(nil)
		}
	}
}

// Return the name of the C function if this is a CGo wrapper. Otherwise, return
// a zero-length string.
func (f *Function) CName() string {
	name := f.Name()
	if strings.HasPrefix(name, "_Cfunc_") {
		// emitted by `go tool cgo`
		return name[len("_Cfunc_"):]
	}
	if strings.HasPrefix(name, "C.") {
		// created by ../loader/cgo.go
		return name[2:]
	}
	return ""
}

// Get all methods of a type.
func getAllMethods(prog *ssa.Program, typ types.Type) []*types.Selection {
	ms := prog.MethodSets.MethodSet(typ)
	methods := make([]*types.Selection, ms.Len())
	for i := 0; i < ms.Len(); i++ {
		methods[i] = ms.At(i)
	}
	return methods
}

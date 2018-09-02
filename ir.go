package main

import (
	"go/ast"
	"go/types"
	"sort"
	"strings"

	"github.com/aykevl/llvm/bindings/go/llvm"
	"golang.org/x/tools/go/ssa"
)

// This file provides a wrapper around go/ssa values and adds extra
// functionality to them.

// View on all functions, types, and globals in a program, with analysis
// results.
type Program struct {
	program              *ssa.Program
	mainPkg              *ssa.Package
	Functions            []*Function
	functionMap          map[*ssa.Function]*Function
	Globals              []*Global
	globalMap            map[*ssa.Global]*Global
	NamedTypes           []*NamedType
	needsScheduler       bool
	goCalls              []*ssa.Go
	typesWithMethods     map[string]*InterfaceType // see AnalyseInterfaceConversions
	typesWithoutMethods  map[string]int            // see AnalyseInterfaceConversions
	methodSignatureNames map[string]int
	fpWithContext        map[string]struct{} // see AnalyseFunctionPointers
}

// Function or method.
type Function struct {
	fn           *ssa.Function
	llvmFn       llvm.Value
	linkName     string
	blocking     bool
	flag         bool        // used by dead code elimination
	addressTaken bool        // used as function pointer, calculated by AnalyseFunctionPointers
	parents      []*Function // calculated by AnalyseCallgraph
	children     []*Function // calculated by AnalyseCallgraph
}

// Global variable, possibly constant.
type Global struct {
	g           *ssa.Global
	llvmGlobal  llvm.Value
	flag        bool // used by dead code elimination
	initializer Value
}

// Type with a name and possibly methods.
type NamedType struct {
	t        *ssa.Type
	llvmType llvm.Type
}

// Type that is at some point put in an interface.
type InterfaceType struct {
	t       types.Type
	Num     int
	Methods map[string]*types.Selection
}

// Create and intialize a new *Program from a *ssa.Program.
func NewProgram(program *ssa.Program, mainPath string) *Program {
	return &Program{
		program:              program,
		mainPkg:              program.ImportedPackage(mainPath),
		functionMap:          make(map[*ssa.Function]*Function),
		globalMap:            make(map[*ssa.Global]*Global),
		methodSignatureNames: make(map[string]int),
	}
}

// Add a package to this Program. All packages need to be added first before any
// analysis is done for correct results.
func (p *Program) AddPackage(pkg *ssa.Package) {
	memberNames := make([]string, 0)
	for name := range pkg.Members {
		if isCGoInternal(name) {
			continue
		}
		memberNames = append(memberNames, name)
	}
	sort.Strings(memberNames)

	for _, name := range memberNames {
		member := pkg.Members[name]
		switch member := member.(type) {
		case *ssa.Function:
			if isCGoInternal(member.Name()) {
				continue
			}
			p.addFunction(member)
		case *ssa.Type:
			t := &NamedType{t: member}
			p.NamedTypes = append(p.NamedTypes, t)
			methods := getAllMethods(pkg.Prog, member.Type())
			if !types.IsInterface(member.Type()) {
				// named type
				for _, method := range methods {
					p.addFunction(pkg.Prog.MethodValue(method))
				}
			}
		case *ssa.Global:
			g := &Global{g: member}
			p.Globals = append(p.Globals, g)
			p.globalMap[member] = g
		case *ssa.NamedConst:
			// Ignore: these are already resolved.
		default:
			panic("unknown member type: " + member.String())
		}
	}
}

func (p *Program) addFunction(ssaFn *ssa.Function) {
	f := &Function{fn: ssaFn}
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

func (p *Program) GetGlobal(ssaGlobal *ssa.Global) *Global {
	return p.globalMap[ssaGlobal]
}

// Parse compiler directives in the preceding comments.
func (f *Function) parsePragmas() {
	if f.fn.Syntax() == nil {
		return
	}
	if decl, ok := f.fn.Syntax().(*ast.FuncDecl); ok && decl.Doc != nil {
		for _, comment := range decl.Doc.List {
			if !strings.HasPrefix(comment.Text, "//go:") {
				continue
			}
			parts := strings.Fields(comment.Text)
			switch parts[0] {
			case "//go:linkname":
				if len(parts) != 3 || parts[1] != f.fn.Name() {
					continue
				}
				// Only enable go:linkname when the package imports "unsafe".
				// This is a slightly looser requirement than what gc uses: gc
				// requires the file to import "unsafe", not the package as a
				// whole.
				if hasUnsafeImport(f.fn.Pkg.Pkg) {
					f.linkName = parts[2]
				}
			}
		}
	}
}

// Return the link name for this function.
func (f *Function) LinkName() string {
	if f.linkName != "" {
		return f.linkName
	}
	if f.fn.Signature.Recv() != nil {
		// Method on a defined type (which may be a pointer).
		return f.fn.RelString(nil)
	} else {
		// Bare function.
		if name := f.CName(); name != "" {
			// Name CGo functions directly.
			return name
		} else {
			return f.fn.RelString(nil)
		}
	}
}

// Return the name of the C function if this is a CGo wrapper. Otherwise, return
// a zero-length string.
func (f *Function) CName() string {
	name := f.fn.Name()
	if strings.HasPrefix(name, "_Cfunc_") {
		return name[len("_Cfunc_"):]
	}
	return ""
}

// Return the link name for this global.
func (g *Global) LinkName() string {
	if strings.HasPrefix(g.g.Name(), "_extern_") {
		return g.g.Name()[len("_extern_"):]
	} else {
		return g.g.RelString(nil)
	}
}

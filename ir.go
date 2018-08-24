package main

import (
	"go/types"
	"sort"
	"strings"

	"github.com/aykevl/llvm/bindings/go/llvm"
	"golang.org/x/tools/go/ssa"
)

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
	typesWithMethods     map[string]*InterfaceType
	typesWithoutMethods  map[string]int
	methodSignatureNames map[string]int
}

// Function or method.
type Function struct {
	fn       *ssa.Function
	llvmFn   llvm.Value
	blocking bool
	flag     bool        // used by dead code elimination
	parents  []*Function // calculated by AnalyseCallgraph
	children []*Function
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
	p.Functions = append(p.Functions, f)
	p.functionMap[ssaFn] = f
}

func (p *Program) GetFunction(ssaFn *ssa.Function) *Function {
	return p.functionMap[ssaFn]
}

func (p *Program) GetGlobal(ssaGlobal *ssa.Global) *Global {
	return p.globalMap[ssaGlobal]
}

// Return the link name for this function.
func (f *Function) LinkName(blocking bool) string {
	suffix := ""
	if blocking {
		suffix = "$async"
	}
	if f.fn.Signature.Recv() != nil {
		// Method on a defined type (which may be a pointer).
		return f.fn.RelString(nil) + suffix
	} else {
		// Bare function.
		if name := f.CName(); name != "" {
			// Name CGo functions directly.
			return name
		} else {
			name := f.fn.RelString(nil) + suffix
			if f.fn.Pkg.Pkg.Path() == "runtime" && strings.HasPrefix(f.fn.Name(), "_llvm_") {
				// Special case for LLVM intrinsics in the runtime.
				name = "llvm." + strings.Replace(f.fn.Name()[len("_llvm_"):], "_", ".", -1)
			}
			return name
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

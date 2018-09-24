package ir

import (
	"go/ast"
	"go/token"
	"go/types"
	"sort"
	"strings"

	"github.com/aykevl/llvm/bindings/go/llvm"
	"golang.org/x/tools/go/loader"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
)

// This file provides a wrapper around go/ssa values and adds extra
// functionality to them.

// View on all functions, types, and globals in a program, with analysis
// results.
type Program struct {
	Program              *ssa.Program
	mainPkg              *ssa.Package
	Functions            []*Function
	functionMap          map[*ssa.Function]*Function
	Globals              []*Global
	globalMap            map[*ssa.Global]*Global
	comments             map[string]*ast.CommentGroup
	NamedTypes           []*NamedType
	needsScheduler       bool
	goCalls              []*ssa.Go
	typesWithMethods     map[string]*TypeWithMethods // see AnalyseInterfaceConversions
	typesWithoutMethods  map[string]int              // see AnalyseInterfaceConversions
	methodSignatureNames map[string]int              // see MethodNum
	interfaces           map[string]*Interface       // see AnalyseInterfaceConversions
	fpWithContext        map[string]struct{}         // see AnalyseFunctionPointers
}

// Function or method.
type Function struct {
	*ssa.Function
	LLVMFn       llvm.Value
	linkName     string      // go:linkname or go:export pragma
	exported     bool        // go:export
	nobounds     bool        // go:nobounds pragma
	blocking     bool        // calculated by AnalyseBlockingRecursive
	flag         bool        // used by dead code elimination
	addressTaken bool        // used as function pointer, calculated by AnalyseFunctionPointers
	parents      []*Function // calculated by AnalyseCallgraph
	children     []*Function // calculated by AnalyseCallgraph
}

// Global variable, possibly constant.
type Global struct {
	*ssa.Global
	program     *Program
	LLVMGlobal  llvm.Value
	linkName    string // go:extern
	extern      bool   // go:extern
	initializer Value
}

// Type with a name and possibly methods.
type NamedType struct {
	*ssa.Type
	LLVMType llvm.Type
}

// Type that is at some point put in an interface.
type TypeWithMethods struct {
	t       types.Type
	Num     int
	Methods map[string]*types.Selection
}

// Interface type that is at some point used in a type assert (to check whether
// it implements another interface).
type Interface struct {
	Num  int
	Type *types.Interface
}

// Create and intialize a new *Program from a *ssa.Program.
func NewProgram(lprogram *loader.Program, mainPath string) *Program {
	comments := map[string]*ast.CommentGroup{}
	for _, pkgInfo := range lprogram.AllPackages {
		for _, file := range pkgInfo.Files {
			for _, decl := range file.Decls {
				switch decl := decl.(type) {
				case *ast.GenDecl:
					switch decl.Tok {
					case token.VAR:
						if len(decl.Specs) != 1 {
							continue
						}
						for _, spec := range decl.Specs {
							valueSpec := spec.(*ast.ValueSpec)
							for _, valueName := range valueSpec.Names {
								id := pkgInfo.Pkg.Path() + "." + valueName.Name
								comments[id] = decl.Doc
							}
						}
					}
				}
			}
		}
	}

	program := ssautil.CreateProgram(lprogram, ssa.SanityCheckFunctions|ssa.BareInits|ssa.GlobalDebug)
	program.Build()

	// Make a list of packages in import order.
	packageList := []*ssa.Package{}
	packageSet := map[string]struct{}{}
	worklist := []string{"runtime", mainPath}
	for len(worklist) != 0 {
		pkgPath := worklist[0]
		pkg := program.ImportedPackage(pkgPath)
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

	p := &Program{
		Program:              program,
		mainPkg:              program.ImportedPackage(mainPath),
		functionMap:          make(map[*ssa.Function]*Function),
		globalMap:            make(map[*ssa.Global]*Global),
		methodSignatureNames: make(map[string]int),
		interfaces:           make(map[string]*Interface),
		comments:             comments,
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
			t := &NamedType{Type: member}
			p.NamedTypes = append(p.NamedTypes, t)
			methods := getAllMethods(pkg.Prog, member.Type())
			if !types.IsInterface(member.Type()) {
				// named type
				for _, method := range methods {
					p.addFunction(pkg.Prog.MethodValue(method))
				}
			}
		case *ssa.Global:
			g := &Global{program: p, Global: member}
			doc := p.comments[g.RelString(nil)]
			if doc != nil {
				g.parsePragmas(doc)
			}
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

func (p *Program) GetGlobal(ssaGlobal *ssa.Global) *Global {
	return p.globalMap[ssaGlobal]
}

// SortMethods sorts the list of methods by method ID.
func (p *Program) SortMethods(methods []*types.Selection) {
	m := &methodList{methods: methods, program: p}
	sort.Sort(m)
}

// SortFuncs sorts the list of functions by method ID.
func (p *Program) SortFuncs(funcs []*types.Func) {
	m := &funcList{funcs: funcs, program: p}
	sort.Sort(m)
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
			if !strings.HasPrefix(comment.Text, "//go:") {
				continue
			}
			parts := strings.Fields(comment.Text)
			switch parts[0] {
			case "//go:export":
				if len(parts) != 2 {
					continue
				}
				f.linkName = parts[1]
				f.exported = true
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
	return f.exported
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
		return name[len("_Cfunc_"):]
	}
	return ""
}

// Parse //go: pragma comments from the source.
func (g *Global) parsePragmas(doc *ast.CommentGroup) {
	for _, comment := range doc.List {
		if !strings.HasPrefix(comment.Text, "//go:") {
			continue
		}
		parts := strings.Fields(comment.Text)
		switch parts[0] {
		case "//go:extern":
			g.extern = true
			if len(parts) == 2 {
				g.linkName = parts[1]
			}
		}
	}
}

// Return the link name for this global.
func (g *Global) LinkName() string {
	if g.linkName != "" {
		return g.linkName
	}
	return g.RelString(nil)
}

func (g *Global) IsExtern() bool {
	return g.extern
}

func (g *Global) Initializer() Value {
	return g.initializer
}

// Wrapper type to implement sort.Interface for []*types.Selection.
type methodList struct {
	methods []*types.Selection
	program *Program
}

func (m *methodList) Len() int {
	return len(m.methods)
}

func (m *methodList) Less(i, j int) bool {
	iid := m.program.MethodNum(m.methods[i].Obj().(*types.Func))
	jid := m.program.MethodNum(m.methods[j].Obj().(*types.Func))
	return iid < jid
}

func (m *methodList) Swap(i, j int) {
	m.methods[i], m.methods[j] = m.methods[j], m.methods[i]
}

// Wrapper type to implement sort.Interface for []*types.Func.
type funcList struct {
	funcs   []*types.Func
	program *Program
}

func (fl *funcList) Len() int {
	return len(fl.funcs)
}

func (fl *funcList) Less(i, j int) bool {
	iid := fl.program.MethodNum(fl.funcs[i])
	jid := fl.program.MethodNum(fl.funcs[j])
	return iid < jid
}

func (fl *funcList) Swap(i, j int) {
	fl.funcs[i], fl.funcs[j] = fl.funcs[j], fl.funcs[i]
}

// Return true if this is a CGo-internal function that can be ignored.
func isCGoInternal(name string) bool {
	if strings.HasPrefix(name, "_Cgo_") || strings.HasPrefix(name, "_cgo") {
		// _Cgo_ptr, _Cgo_use, _cgoCheckResult, _cgo_runtime_cgocall
		return true // CGo-internal functions
	}
	if strings.HasPrefix(name, "__cgofn__cgo_") {
		return true // CGo function pointer in global scope
	}
	return false
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

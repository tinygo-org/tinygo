package loader

import (
	"bytes"
	"errors"
	"go/ast"
	"go/build"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"sort"
	"strings"
	"text/template"

	"github.com/tinygo-org/tinygo/cgo"
)

// Program holds all packages and some metadata about the program as a whole.
type Program struct {
	mainPkg      string
	Build        *build.Context
	OverlayBuild *build.Context
	OverlayPath  func(path string) string
	Packages     map[string]*Package
	sorted       []*Package
	fset         *token.FileSet
	TypeChecker  types.Config
	Dir          string // current working directory (for error reporting)
	TINYGOROOT   string // root of the TinyGo installation or root of the source code
	CFlags       []string
	LDFlags      []string
	ClangHeaders string
}

// Package holds a loaded package, its imports, and its parsed files.
type Package struct {
	*Program
	*build.Package
	Imports   map[string]*Package
	Importing bool
	Files     []*ast.File
	Pkg       *types.Package
	types.Info
}

// Import loads the given package relative to srcDir (for the vendor directory).
// It only loads the current package without recursion.
func (p *Program) Import(path, srcDir string, pos token.Position) (*Package, error) {
	if p.Packages == nil {
		p.Packages = make(map[string]*Package)
	}

	// Load this package.
	ctx := p.Build
	if newPath := p.OverlayPath(path); newPath != "" {
		ctx = p.OverlayBuild
		path = newPath
	}
	buildPkg, err := ctx.Import(path, srcDir, build.ImportComment)
	if err != nil {
		return nil, scanner.Error{
			Pos: pos,
			Msg: err.Error(), // TODO: define a new error type that will wrap the inner error
		}
	}
	if existingPkg, ok := p.Packages[buildPkg.ImportPath]; ok {
		// Already imported, or at least started the import.
		return existingPkg, nil
	}
	p.sorted = nil // invalidate the sorted order of packages
	pkg := p.newPackage(buildPkg)
	p.Packages[buildPkg.ImportPath] = pkg

	if p.mainPkg == "" {
		p.mainPkg = buildPkg.ImportPath
	}

	return pkg, nil
}

// ImportFile loads and parses the import statements in the given path and
// creates a pseudo-package out of it.
func (p *Program) ImportFile(path string) (*Package, error) {
	if p.Packages == nil {
		p.Packages = make(map[string]*Package)
	}
	if _, ok := p.Packages[path]; ok {
		// unlikely
		return nil, errors.New("loader: cannot import file that is already imported as package: " + path)
	}

	file, err := p.parseFile(path, parser.ImportsOnly)
	if err != nil {
		return nil, err
	}
	buildPkg := &build.Package{
		Dir:        filepath.Dir(path),
		ImportPath: path,
		GoFiles:    []string{filepath.Base(path)},
	}
	for _, importSpec := range file.Imports {
		buildPkg.Imports = append(buildPkg.Imports, importSpec.Path.Value[1:len(importSpec.Path.Value)-1])
	}
	p.sorted = nil // invalidate the sorted order of packages
	pkg := p.newPackage(buildPkg)
	p.Packages[buildPkg.ImportPath] = pkg

	if p.mainPkg == "" {
		p.mainPkg = buildPkg.ImportPath
	}

	return pkg, nil
}

// newPackage instantiates a new *Package object with initialized members.
func (p *Program) newPackage(pkg *build.Package) *Package {
	return &Package{
		Program: p,
		Package: pkg,
		Imports: make(map[string]*Package, len(pkg.Imports)),
		Info: types.Info{
			Types:      make(map[ast.Expr]types.TypeAndValue),
			Defs:       make(map[*ast.Ident]types.Object),
			Uses:       make(map[*ast.Ident]types.Object),
			Implicits:  make(map[ast.Node]types.Object),
			Scopes:     make(map[ast.Node]*types.Scope),
			Selections: make(map[*ast.SelectorExpr]*types.Selection),
		},
	}
}

// Sorted returns a list of all packages, sorted in a way that no packages come
// before the packages they depend upon.
func (p *Program) Sorted() []*Package {
	if p.sorted == nil {
		p.sort()
	}
	return p.sorted
}

func (p *Program) sort() {
	p.sorted = nil
	packageList := make([]*Package, 0, len(p.Packages))
	packageSet := make(map[string]struct{}, len(p.Packages))
	worklist := make([]string, 0, len(p.Packages))
	for path := range p.Packages {
		worklist = append(worklist, path)
	}
	sort.Strings(worklist)
	for len(worklist) != 0 {
		pkgPath := worklist[0]
		pkg := p.Packages[pkgPath]

		if _, ok := packageSet[pkgPath]; ok {
			// Package already in the final package list.
			worklist = worklist[1:]
			continue
		}

		unsatisfiedImports := make([]string, 0)
		for _, pkg := range pkg.Imports {
			if _, ok := packageSet[pkg.ImportPath]; ok {
				continue
			}
			unsatisfiedImports = append(unsatisfiedImports, pkg.ImportPath)
		}
		sort.Strings(unsatisfiedImports)
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

	p.sorted = packageList
}

// Parse recursively imports all packages, parses them, and typechecks them.
//
// The returned error may be an Errors error, which contains a list of errors.
//
// Idempotent.
func (p *Program) Parse(compileTestBinary bool) error {
	includeTests := compileTestBinary

	// Load all imports
	for _, pkg := range p.Sorted() {
		err := pkg.importRecursively(includeTests)
		if err != nil {
			if err, ok := err.(*ImportCycleError); ok {
				if pkg.ImportPath != err.Packages[0] {
					err.Packages = append([]string{pkg.ImportPath}, err.Packages...)
				}
			}
			return err
		}
	}

	// Parse all packages.
	for _, pkg := range p.Sorted() {
		err := pkg.Parse(includeTests)
		if err != nil {
			return err
		}
	}

	if compileTestBinary {
		err := p.SwapTestMain()
		if err != nil {
			return err
		}
	}

	// Typecheck all packages.
	for _, pkg := range p.Sorted() {
		err := pkg.Check()
		if err != nil {
			return err
		}
	}

	return nil
}

func (p *Program) SwapTestMain() error {
	var tests []string

	isTestFunc := func(f *ast.FuncDecl) bool {
		// TODO: improve signature check
		if strings.HasPrefix(f.Name.Name, "Test") && f.Name.Name != "TestMain" {
			return true
		}
		return false
	}
	mainPkg := p.Packages[p.mainPkg]
	for _, f := range mainPkg.Files {
		for i, d := range f.Decls {
			switch v := d.(type) {
			case *ast.FuncDecl:
				if isTestFunc(v) {
					tests = append(tests, v.Name.Name)
				}
				if v.Name.Name == "main" {
					// Remove main
					if len(f.Decls) == 1 {
						f.Decls = make([]ast.Decl, 0)
					} else {
						f.Decls[i] = f.Decls[len(f.Decls)-1]
						f.Decls = f.Decls[:len(f.Decls)-1]
					}
				}
			}
		}
	}

	// TODO: Check if they defined a TestMain and call it instead of testing.TestMain
	const mainBody = `package main

import (
	"testing"
)

func main () {
	m := &testing.M{
		Tests: []testing.TestToCall{
{{range .TestFunctions}}
			{Name: "{{.}}", Func: {{.}}},
{{end}}
		},
	}

	testing.TestMain(m)
}
`
	tmpl := template.Must(template.New("testmain").Parse(mainBody))
	b := bytes.Buffer{}
	tmplData := struct {
		TestFunctions []string
	}{
		TestFunctions: tests,
	}

	err := tmpl.Execute(&b, tmplData)
	if err != nil {
		return err
	}
	path := filepath.Join(p.mainPkg, "$testmain.go")

	if p.fset == nil {
		p.fset = token.NewFileSet()
	}

	newMain, err := parser.ParseFile(p.fset, path, b.Bytes(), parser.AllErrors)
	if err != nil {
		return err
	}
	mainPkg.Files = append(mainPkg.Files, newMain)

	return nil
}

// parseFile is a wrapper around parser.ParseFile.
func (p *Program) parseFile(path string, mode parser.Mode) (*ast.File, error) {
	if p.fset == nil {
		p.fset = token.NewFileSet()
	}

	rd, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer rd.Close()
	relpath := path
	if filepath.IsAbs(path) {
		rp, err := filepath.Rel(p.Dir, path)
		if err == nil {
			relpath = rp
		}
	}
	return parser.ParseFile(p.fset, relpath, rd, mode)
}

// Parse parses and typechecks this package.
//
// Idempotent.
func (p *Package) Parse(includeTests bool) error {
	if len(p.Files) != 0 {
		return nil
	}

	// Load the AST.
	// TODO: do this in parallel.
	if p.ImportPath == "unsafe" {
		// Special case for the unsafe package. Don't even bother loading
		// the files.
		p.Pkg = types.Unsafe
		return nil
	}

	files, err := p.parseFiles(includeTests)
	if err != nil {
		return err
	}
	p.Files = files

	return nil
}

// Check runs the package through the typechecker. The package must already be
// loaded and all dependencies must have been checked already.
//
// Idempotent.
func (p *Package) Check() error {
	if p.Pkg != nil {
		return nil
	}

	var typeErrors []error
	checker := p.TypeChecker
	checker.Error = func(err error) {
		typeErrors = append(typeErrors, err)
	}

	// Do typechecking of the package.
	checker.Importer = p

	typesPkg, err := checker.Check(p.ImportPath, p.fset, p.Files, &p.Info)
	if err != nil {
		if err, ok := err.(Errors); ok {
			return err
		}
		return Errors{p, typeErrors}
	}
	p.Pkg = typesPkg
	return nil
}

// parseFiles parses the loaded list of files and returns this list.
func (p *Package) parseFiles(includeTests bool) ([]*ast.File, error) {
	// TODO: do this concurrently.
	var files []*ast.File
	var fileErrs []error

	var gofiles []string
	if includeTests {
		gofiles = make([]string, 0, len(p.GoFiles)+len(p.TestGoFiles))
		gofiles = append(gofiles, p.GoFiles...)
		gofiles = append(gofiles, p.TestGoFiles...)
	} else {
		gofiles = p.GoFiles
	}

	for _, file := range gofiles {
		f, err := p.parseFile(filepath.Join(p.Package.Dir, file), parser.ParseComments)
		if err != nil {
			fileErrs = append(fileErrs, err)
			continue
		}
		if err != nil {
			fileErrs = append(fileErrs, err)
			continue
		}
		files = append(files, f)
	}
	for _, file := range p.CgoFiles {
		path := filepath.Join(p.Package.Dir, file)
		f, err := p.parseFile(path, parser.ParseComments)
		if err != nil {
			fileErrs = append(fileErrs, err)
			continue
		}
		files = append(files, f)
	}
	if len(p.CgoFiles) != 0 {
		cflags := append(p.CFlags, "-I"+p.Package.Dir)
		if p.ClangHeaders != "" {
			cflags = append(cflags, "-Xclang", "-internal-isystem", "-Xclang", p.ClangHeaders)
		}
		generated, ldflags, errs := cgo.Process(files, p.Program.Dir, p.fset, cflags)
		if errs != nil {
			fileErrs = append(fileErrs, errs...)
		}
		files = append(files, generated)
		p.LDFlags = append(p.LDFlags, ldflags...)
	}
	if len(fileErrs) != 0 {
		return nil, Errors{p, fileErrs}
	}

	return files, nil
}

// Import implements types.Importer. It loads and parses packages it encounters
// along the way, if needed.
func (p *Package) Import(to string) (*types.Package, error) {
	if to == "unsafe" {
		return types.Unsafe, nil
	}
	if _, ok := p.Imports[to]; ok {
		return p.Imports[to].Pkg, nil
	} else {
		return nil, errors.New("package not imported: " + to)
	}
}

// importRecursively calls Program.Import() on all imported packages, and calls
// importRecursively() on the imported packages as well.
//
// Idempotent.
func (p *Package) importRecursively(includeTests bool) error {
	p.Importing = true

	imports := p.Package.Imports
	if includeTests {
		imports = append(imports, p.Package.TestImports...)
	}

	for _, to := range imports {
		if to == "C" {
			// Do CGo processing in a later stage.
			continue
		}
		if _, ok := p.Imports[to]; ok {
			continue
		}
		// Find error location.
		var pos token.Position
		if len(p.Package.ImportPos[to]) > 0 {
			pos = p.Package.ImportPos[to][0]
		} else {
			pos = token.Position{Filename: p.Package.ImportPath}
		}
		importedPkg, err := p.Program.Import(to, p.Package.Dir, pos)
		if err != nil {
			if err, ok := err.(*ImportCycleError); ok {
				err.Packages = append([]string{p.ImportPath}, err.Packages...)
			}
			return err
		}
		if importedPkg.Importing {
			return &ImportCycleError{[]string{p.ImportPath, importedPkg.ImportPath}, p.ImportPos[to]}
		}
		err = importedPkg.importRecursively(false)
		if err != nil {
			if err, ok := err.(*ImportCycleError); ok {
				err.Packages = append([]string{p.ImportPath}, err.Packages...)
			}
			return err
		}
		p.Imports[to] = importedPkg
	}
	p.Importing = false
	return nil
}

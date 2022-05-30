package loader

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"

	"github.com/tinygo-org/tinygo/cgo"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

// Program holds all packages and some metadata about the program as a whole.
type Program struct {
	config       *compileopts.Config
	clangHeaders string
	typeChecker  types.Config
	goroot       string // synthetic GOROOT
	workingDir   string

	Packages map[string]*Package
	sorted   []*Package
	fset     *token.FileSet

	// Information obtained during parsing.
	LDFlags []string
}

// PackageJSON is a subset of the JSON struct returned from `go list`.
type PackageJSON struct {
	Dir        string
	ImportPath string
	Name       string
	ForTest    string
	Root       string
	Module     struct {
		Path      string
		Main      bool
		Dir       string
		GoMod     string
		GoVersion string
	}

	// Source files
	GoFiles  []string
	CgoFiles []string
	CFiles   []string

	// Dependency information
	Imports   []string
	ImportMap map[string]string

	// Error information
	Error *struct {
		ImportStack []string
		Pos         string
		Err         string
	}
}

// Package holds a loaded package, its imports, and its parsed files.
type Package struct {
	PackageJSON

	program    *Program
	Files      []*ast.File
	FileHashes map[string][]byte
	CFlags     []string // CFlags used during CGo preprocessing (only set if CGo is used)
	CGoHeaders []string // text above 'import "C"' lines
	Pkg        *types.Package
	info       types.Info
}

// Load loads the given package with all dependencies (including the runtime
// package). Call .Parse() afterwards to parse all Go files (including CGo
// processing, if necessary).
func Load(config *compileopts.Config, inputPkgs []string, clangHeaders string, typeChecker types.Config) (*Program, error) {
	goroot, err := GetCachedGoroot(config)
	if err != nil {
		return nil, err
	}
	var wd string
	if config.Options.Directory != "" {
		wd = config.Options.Directory
	} else {
		wd, err = os.Getwd()
		if err != nil {
			return nil, err
		}
	}
	p := &Program{
		config:       config,
		clangHeaders: clangHeaders,
		typeChecker:  typeChecker,
		goroot:       goroot,
		workingDir:   wd,
		Packages:     make(map[string]*Package),
		fset:         token.NewFileSet(),
	}

	// List the dependencies of this package, in raw JSON format.
	extraArgs := []string{"-json", "-deps"}
	if config.TestConfig.CompileTestBinary {
		extraArgs = append(extraArgs, "-test")
	}
	cmd, err := List(config, extraArgs, inputPkgs)
	if err != nil {
		return nil, err
	}
	buf := &bytes.Buffer{}
	cmd.Stdout = buf
	cmd.Stderr = os.Stderr
	err = cmd.Run()
	if err != nil {
		if exitErr, ok := err.(*exec.ExitError); ok {
			os.Exit(exitErr.ExitCode())
		}
		return nil, fmt.Errorf("failed to run `go list`: %s", err)
	}

	// Parse the returned json from `go list`.
	decoder := json.NewDecoder(buf)
	for {
		pkg := &Package{
			program:    p,
			FileHashes: make(map[string][]byte),
			info: types.Info{
				Types:      make(map[ast.Expr]types.TypeAndValue),
				Defs:       make(map[*ast.Ident]types.Object),
				Uses:       make(map[*ast.Ident]types.Object),
				Implicits:  make(map[ast.Node]types.Object),
				Scopes:     make(map[ast.Node]*types.Scope),
				Selections: make(map[*ast.SelectorExpr]*types.Selection),
			},
		}
		err := decoder.Decode(&pkg.PackageJSON)
		if err != nil {
			if err == io.EOF {
				break
			}
			return nil, err
		}
		if pkg.Error != nil {
			// There was an error while importing (for example, a circular
			// dependency).
			pos := token.Position{}
			fields := strings.Split(pkg.Error.Pos, ":")
			if len(fields) >= 2 {
				// There is some file/line/column information.
				if n, err := strconv.Atoi(fields[len(fields)-2]); err == nil {
					// Format: filename.go:line:colum
					pos.Filename = strings.Join(fields[:len(fields)-2], ":")
					pos.Line = n
					pos.Column, _ = strconv.Atoi(fields[len(fields)-1])
				} else {
					// Format: filename.go:line
					pos.Filename = strings.Join(fields[:len(fields)-1], ":")
					pos.Line, _ = strconv.Atoi(fields[len(fields)-1])
				}
				pos.Filename = p.getOriginalPath(pos.Filename)
			}
			err := scanner.Error{
				Pos: pos,
				Msg: pkg.Error.Err,
			}
			if len(pkg.Error.ImportStack) != 0 {
				return nil, Error{
					ImportStack: pkg.Error.ImportStack,
					Err:         err,
				}
			}
			return nil, err
		}
		if config.TestConfig.CompileTestBinary {
			// When creating a test binary, `go list` will list two or three
			// packages used for testing the package. The first is the original
			// package as if it were built normally, the second is the same
			// package but with the *_test.go files included. A possible third
			// may be included for _test packages (such as math_test), used to
			// test the external API with no access to internal functions.
			// All packages that are necessary for testing (including the to be
			// tested package with *_test.go files, but excluding the original
			// unmodified package) have a suffix added to the import path, for
			// example the math package has import path "math [math.test]" and
			// test dependencies such as fmt will have an import path of the
			// form "fmt [math.test]".
			// The code below removes this suffix, and if this results in a
			// duplicate (which happens with the to-be-tested package without
			// *.test.go files) the previous package is removed from the list of
			// packages included in this build.
			// This is necessary because the change in import paths results in
			// breakage to //go:linkname. Additionally, the duplicated package
			// slows down the build and so is best removed.
			if pkg.ForTest != "" && strings.HasSuffix(pkg.ImportPath, " ["+pkg.ForTest+".test]") {
				newImportPath := pkg.ImportPath[:len(pkg.ImportPath)-len(" ["+pkg.ForTest+".test]")]
				if _, ok := p.Packages[newImportPath]; ok {
					// Delete the previous package (that this package overrides).
					delete(p.Packages, newImportPath)
					for i, pkg := range p.sorted {
						if pkg.ImportPath == newImportPath {
							p.sorted = append(p.sorted[:i], p.sorted[i+1:]...) // remove element from slice
							break
						}
					}
				}
				pkg.ImportPath = newImportPath
			}
		}
		p.sorted = append(p.sorted, pkg)
		p.Packages[pkg.ImportPath] = pkg
	}

	if config.TestConfig.CompileTestBinary && !strings.HasSuffix(p.sorted[len(p.sorted)-1].ImportPath, ".test") {
		// Trying to compile a test binary but there are no test files in this
		// package.
		return p, NoTestFilesError{p.sorted[len(p.sorted)-1].ImportPath}
	}

	return p, nil
}

// getOriginalPath looks whether this path is in the generated GOROOT and if so,
// replaces the path with the original path (in GOROOT or TINYGOROOT). Otherwise
// the input path is returned.
func (p *Program) getOriginalPath(path string) string {
	originalPath := path
	if strings.HasPrefix(path, p.goroot+string(filepath.Separator)) {
		// If this file is part of the synthetic GOROOT, try to infer the
		// original path.
		relpath := path[len(filepath.Join(p.goroot, "src"))+1:]
		realgorootPath := filepath.Join(goenv.Get("GOROOT"), "src", relpath)
		if _, err := os.Stat(realgorootPath); err == nil {
			originalPath = realgorootPath
		}
		maybeInTinyGoRoot := false
		for prefix := range pathsToOverride(needsSyscallPackage(p.config.BuildTags())) {
			if runtime.GOOS == "windows" {
				prefix = strings.ReplaceAll(prefix, "/", "\\")
			}
			if !strings.HasPrefix(relpath, prefix) {
				continue
			}
			maybeInTinyGoRoot = true
		}
		if maybeInTinyGoRoot {
			tinygoPath := filepath.Join(goenv.Get("TINYGOROOT"), "src", relpath)
			if _, err := os.Stat(tinygoPath); err == nil {
				originalPath = tinygoPath
			}
		}
	}
	return originalPath
}

// Sorted returns a list of all packages, sorted in a way that no packages come
// before the packages they depend upon.
func (p *Program) Sorted() []*Package {
	return p.sorted
}

// MainPkg returns the last package in the Sorted() slice. This is the main
// package of the program.
func (p *Program) MainPkg() *Package {
	return p.sorted[len(p.sorted)-1]
}

// Parse parses all packages and typechecks them.
//
// The returned error may be an Errors error, which contains a list of errors.
//
// Idempotent.
func (p *Program) Parse() error {
	// Parse all packages.
	// TODO: do this in parallel.
	for _, pkg := range p.sorted {
		err := pkg.Parse()
		if err != nil {
			return fmt.Errorf("when parsing the package %q: %w", pkg.ImportPath, err)
		}
	}

	// Typecheck all packages.
	for _, pkg := range p.sorted {
		err := pkg.Check()
		if err != nil {
			return fmt.Errorf("when checking the package %q: %w", pkg.ImportPath, err)
		}
	}

	return nil
}

// OriginalDir returns the real directory name. It is the same as p.Dir except
// that if it is part of the cached GOROOT, its real location is returned.
func (p *Package) OriginalDir() string {
	return strings.TrimSuffix(p.program.getOriginalPath(p.Dir+string(os.PathSeparator)), string(os.PathSeparator))
}

// parseFile is a wrapper around parser.ParseFile.
func (p *Package) parseFile(path string, mode parser.Mode) (*ast.File, error) {
	originalPath := p.program.getOriginalPath(path)
	data, err := ioutil.ReadFile(path)
	if err != nil {
		return nil, err
	}
	sum := sha512.Sum512_224(data)
	p.FileHashes[originalPath] = sum[:]
	return parser.ParseFile(p.program.fset, originalPath, data, mode)
}

// Parse parses and typechecks this package.
//
// Idempotent.
func (p *Package) Parse() error {
	if len(p.Files) != 0 {
		return nil // nothing to do (?)
	}

	// Load the AST.
	if p.ImportPath == "unsafe" {
		// Special case for the unsafe package, which is defined internally by
		// the types package.
		p.Pkg = types.Unsafe
		return nil
	}

	files, err := p.parseFiles()
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
		return nil // already typechecked
	}

	var typeErrors []error
	checker := p.program.typeChecker // make a copy, because it will be modified
	checker.Error = func(err error) {
		typeErrors = append(typeErrors, err)
	}

	// Do typechecking of the package.
	checker.Importer = p

	packageName := p.ImportPath
	if p == p.program.MainPkg() {
		if p.Name != "main" {
			// Sanity check. Should not ever trigger.
			panic("expected main package to have name 'main'")
		}
		packageName = "main"
	}
	typesPkg, err := checker.Check(packageName, p.program.fset, p.Files, &p.info)
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
func (p *Package) parseFiles() ([]*ast.File, error) {
	var files []*ast.File
	var fileErrs []error

	// Parse all files (incuding CgoFiles).
	parseFile := func(file string) {
		if !filepath.IsAbs(file) {
			file = filepath.Join(p.Dir, file)
		}
		f, err := p.parseFile(file, parser.ParseComments)
		if err != nil {
			fileErrs = append(fileErrs, err)
			return
		}
		files = append(files, f)
	}
	for _, file := range p.GoFiles {
		parseFile(file)
	}
	for _, file := range p.CgoFiles {
		parseFile(file)
	}

	// Do CGo processing.
	// This is done when there are any CgoFiles at all. In that case, len(files)
	// should be non-zero. However, if len(GoFiles) == 0 and len(CgoFiles) == 1
	// and there is a syntax error in a CGo file, len(files) may be 0. Don't try
	// to call cgo.Process in that case as it will only cause issues.
	if len(p.CgoFiles) != 0 && len(files) != 0 {
		var initialCFlags []string
		initialCFlags = append(initialCFlags, p.program.config.CFlags()...)
		initialCFlags = append(initialCFlags, "-I"+p.Dir)
		generated, headerCode, cflags, ldflags, accessedFiles, errs := cgo.Process(files, p.program.workingDir, p.program.fset, initialCFlags, p.program.clangHeaders)
		p.CFlags = append(initialCFlags, cflags...)
		p.CGoHeaders = headerCode
		for path, hash := range accessedFiles {
			p.FileHashes[path] = hash
		}
		if errs != nil {
			fileErrs = append(fileErrs, errs...)
		}
		files = append(files, generated)
		p.program.LDFlags = append(p.program.LDFlags, ldflags...)
	}

	// Only return an error after CGo processing, so that errors in parsing and
	// CGo can be reported together.
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
	if newTo, ok := p.ImportMap[to]; ok && !strings.HasSuffix(newTo, ".test]") {
		to = newTo
	}
	if imported, ok := p.program.Packages[to]; ok {
		return imported.Pkg, nil
	} else {
		return nil, errors.New("package not imported: " + to)
	}
}

package loader

import (
	"bytes"
	"crypto/sha512"
	"encoding/json"
	"errors"
	"fmt"
	"go/ast"
	"go/constant"
	"go/parser"
	"go/scanner"
	"go/token"
	"go/types"
	"io"
	"os"
	"os/exec"
	"path"
	"path/filepath"
	"runtime"
	"strconv"
	"strings"
	"unicode"

	"github.com/tinygo-org/tinygo/cgo"
	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
)

var initFileVersions = func(info *types.Info) {}

// Program holds all packages and some metadata about the program as a whole.
type Program struct {
	config      *compileopts.Config
	typeChecker types.Config
	goroot      string // synthetic GOROOT
	workingDir  string

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

	// Embedded files
	EmbedFiles []string

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

	program      *Program
	Files        []*ast.File
	FileHashes   map[string][]byte
	CFlags       []string // CFlags used during CGo preprocessing (only set if CGo is used)
	CGoHeaders   []string // text above 'import "C"' lines
	EmbedGlobals map[string][]*EmbedFile
	Pkg          *types.Package
	info         types.Info
}

type EmbedFile struct {
	Name      string
	Size      uint64
	Hash      string // hash of the file (as a hex string)
	NeedsData bool   // true if this file is embedded as a byte slice
	Data      []byte // contents of this file (only if NeedsData is set)
}

// Load loads the given package with all dependencies (including the runtime
// package). Call .Parse() afterwards to parse all Go files (including CGo
// processing, if necessary).
func Load(config *compileopts.Config, inputPkg string, typeChecker types.Config) (*Program, error) {
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
		config:      config,
		typeChecker: typeChecker,
		goroot:      goroot,
		workingDir:  wd,
		Packages:    make(map[string]*Package),
		fset:        token.NewFileSet(),
	}

	// List the dependencies of this package, in raw JSON format.
	extraArgs := []string{"-json", "-deps", "-e"}
	if config.TestConfig.CompileTestBinary {
		extraArgs = append(extraArgs, "-test")
	}
	cmd, err := List(config, extraArgs, []string{inputPkg})
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
	var pkgErrors []error
	for {
		pkg := &Package{
			program:      p,
			FileHashes:   make(map[string][]byte),
			EmbedGlobals: make(map[string][]*EmbedFile),
			info: types.Info{
				Types:      make(map[ast.Expr]types.TypeAndValue),
				Instances:  make(map[*ast.Ident]types.Instance),
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
					// Format: filename.go:line:column
					pos.Filename = strings.Join(fields[:len(fields)-2], ":")
					pos.Line = n
					pos.Column, _ = strconv.Atoi(fields[len(fields)-1])
				} else {
					// Format: filename.go:line
					pos.Filename = strings.Join(fields[:len(fields)-1], ":")
					pos.Line, _ = strconv.Atoi(fields[len(fields)-1])
				}
				if abs, err := filepath.Abs(pos.Filename); err == nil {
					// Make the path absolute, so that error messages will be
					// prettier (it will be turned back into a relative path
					// when printing the error).
					pos.Filename = abs
				}
				pos.Filename = p.getOriginalPath(pos.Filename)
			}
			err := scanner.Error{
				Pos: pos,
				Msg: pkg.Error.Err,
			}
			if len(pkg.Error.ImportStack) != 0 {
				pkgErrors = append(pkgErrors, Error{
					ImportStack: pkg.Error.ImportStack,
					Err:         err,
				})
				continue
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

	if len(pkgErrors) != 0 {
		// TODO: use errors.Join in Go 1.20.
		return nil, Errors{
			Errs: pkgErrors,
		}
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
		for prefix := range pathsToOverride(p.config.GoMinorVersion, p.config.BuildTags()) {
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
			return err
		}
	}

	// Typecheck all packages.
	for _, pkg := range p.sorted {
		err := pkg.Check()
		if err != nil {
			return err
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
	data, err := os.ReadFile(path)
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

	// Prepare some state used during type checking.
	var typeErrors []error
	checker := p.program.typeChecker // make a copy, because it will be modified
	checker.Error = func(err error) {
		typeErrors = append(typeErrors, err)
	}
	checker.Importer = p
	if p.Module.GoVersion != "" {
		// Setting the Go version for a module makes sure the type checker
		// errors out on language features not supported in that particular
		// version.
		checker.GoVersion = "go" + p.Module.GoVersion
	} else {
		// Version is not known, so use the currently installed Go version.
		// This is needed for `tinygo run` for example.
		// Normally we'd use goenv.GorootVersionString(), but for compatibility
		// with Go 1.20 and below we need a version in the form of "go1.12" (no
		// patch version).
		major, minor, err := goenv.GetGorootVersion()
		if err != nil {
			return err
		}
		checker.GoVersion = fmt.Sprintf("go%d.%d", major, minor)
	}
	initFileVersions(&p.info)

	// Do typechecking of the package.
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

	p.extractEmbedLines(checker.Error)
	if len(typeErrors) != 0 {
		return Errors{p, typeErrors}
	}

	return nil
}

// parseFiles parses the loaded list of files and returns this list.
func (p *Package) parseFiles() ([]*ast.File, error) {
	var files []*ast.File
	var fileErrs []error

	// Parse all files (including CgoFiles).
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
		initialCFlags = append(initialCFlags, p.program.config.CFlags(true)...)
		initialCFlags = append(initialCFlags, "-I"+p.Dir)
		generated, headerCode, cflags, ldflags, accessedFiles, errs := cgo.Process(files, p.program.workingDir, p.ImportPath, p.program.fset, initialCFlags)
		p.CFlags = append(initialCFlags, cflags...)
		p.CGoHeaders = headerCode
		for path, hash := range accessedFiles {
			p.FileHashes[path] = hash
		}
		if errs != nil {
			fileErrs = append(fileErrs, errs...)
		}
		files = append(files, generated...)
		p.program.LDFlags = append(p.program.LDFlags, ldflags...)
	}

	// Only return an error after CGo processing, so that errors in parsing and
	// CGo can be reported together.
	if len(fileErrs) != 0 {
		return nil, Errors{p, fileErrs}
	}

	return files, nil
}

// extractEmbedLines finds all //go:embed lines in the package and matches them
// against EmbedFiles from `go list`.
func (p *Package) extractEmbedLines(addError func(error)) {
	for _, file := range p.Files {
		// Check for an `import "embed"` line at the start of the file.
		// //go:embed lines are only valid if the given file itself imports the
		// embed package. It is not valid if it is only imported in a separate
		// Go file.
		hasEmbed := false
		for _, importSpec := range file.Imports {
			if importSpec.Path.Value == `"embed"` {
				hasEmbed = true
			}
		}

		for _, decl := range file.Decls {
			switch decl := decl.(type) {
			case *ast.GenDecl:
				if decl.Tok != token.VAR {
					continue
				}
				for _, spec := range decl.Specs {
					spec := spec.(*ast.ValueSpec)
					var doc *ast.CommentGroup
					if decl.Lparen == token.NoPos {
						// Plain 'var' declaration, like:
						//   //go:embed hello.txt
						//   var hello string
						doc = decl.Doc
					} else {
						// Bigger 'var' declaration like:
						//   var (
						//       //go:embed hello.txt
						//       hello string
						//   )
						doc = spec.Doc
					}
					if doc == nil {
						continue
					}

					// Look for //go:embed comments.
					var allPatterns []string
					for _, comment := range doc.List {
						if comment.Text != "//go:embed" && !strings.HasPrefix(comment.Text, "//go:embed ") {
							continue
						}
						if !hasEmbed {
							addError(types.Error{
								Fset: p.program.fset,
								Pos:  comment.Pos() + 2,
								Msg:  "//go:embed only allowed in Go files that import \"embed\"",
							})
							// Continue, because otherwise we might run into
							// issues below.
							continue
						}
						patterns, err := p.parseGoEmbed(comment.Text[len("//go:embed"):], comment.Slash)
						if err != nil {
							addError(err)
							continue
						}
						if len(patterns) == 0 {
							addError(types.Error{
								Fset: p.program.fset,
								Pos:  comment.Pos() + 2,
								Msg:  "usage: //go:embed pattern...",
							})
							continue
						}
						for _, pattern := range patterns {
							// Check that the pattern is well-formed.
							// It must be valid: the Go toolchain has already
							// checked for invalid patterns. But let's check
							// anyway to be sure.
							if _, err := path.Match(pattern, ""); err != nil {
								addError(types.Error{
									Fset: p.program.fset,
									Pos:  comment.Pos(),
									Msg:  "invalid pattern syntax",
								})
								continue
							}
							allPatterns = append(allPatterns, pattern)
						}
					}

					if len(allPatterns) != 0 {
						// This is a //go:embed global. Do a few more checks.
						if len(spec.Names) != 1 {
							addError(types.Error{
								Fset: p.program.fset,
								Pos:  spec.Names[1].NamePos,
								Msg:  "//go:embed cannot apply to multiple vars",
							})
						}
						if spec.Values != nil {
							addError(types.Error{
								Fset: p.program.fset,
								Pos:  spec.Values[0].Pos(),
								Msg:  "//go:embed cannot apply to var with initializer",
							})
						}
						globalName := spec.Names[0].Name
						globalType := p.Pkg.Scope().Lookup(globalName).Type()
						valid, byteSlice := isValidEmbedType(globalType)
						if !valid {
							addError(types.Error{
								Fset: p.program.fset,
								Pos:  spec.Type.Pos(),
								Msg:  "//go:embed cannot apply to var of type " + globalType.String(),
							})
						}

						// Match all //go:embed patterns against the embed files
						// provided by `go list`.
						for _, name := range p.EmbedFiles {
							for _, pattern := range allPatterns {
								if matchPattern(pattern, name) {
									p.EmbedGlobals[globalName] = append(p.EmbedGlobals[globalName], &EmbedFile{
										Name:      name,
										NeedsData: byteSlice,
									})
									break
								}
							}
						}
					}
				}
			}
		}
	}
}

// matchPattern returns true if (and only if) the given pattern would match the
// filename. The pattern could also match a parent directory of name, in which
// case hidden files do not match.
func matchPattern(pattern, name string) bool {
	// Match this file.
	matched, _ := path.Match(pattern, name)
	if matched {
		return true
	}

	// Match parent directories.
	dir := name
	for {
		dir, _ = path.Split(dir)
		if dir == "" {
			return false
		}
		dir = path.Clean(dir)
		if matched, _ := path.Match(pattern, dir); matched {
			// Pattern matches the directory.
			suffix := name[len(dir):]
			if strings.Contains(suffix, "/_") || strings.Contains(suffix, "/.") {
				// Pattern matches a hidden file.
				// Hidden files are included when listed directly as a
				// pattern, but not when they are part of a directory tree.
				// Source:
				// > If a pattern names a directory, all files in the
				// > subtree rooted at that directory are embedded
				// > (recursively), except that files with names beginning
				// > with ‘.’ or ‘_’ are excluded.
				return false
			}
			return true
		}
	}
}

// parseGoEmbed is like strings.Fields but for a //go:embed line. It parses
// regular fields and quoted fields (that may contain spaces).
func (p *Package) parseGoEmbed(args string, pos token.Pos) (patterns []string, err error) {
	args = strings.TrimSpace(args)
	initialLen := len(args)
	for args != "" {
		patternPos := pos + token.Pos(initialLen-len(args))
		switch args[0] {
		case '`', '"', '\\':
			// Parse the next pattern using the Go scanner.
			// This is perhaps a bit overkill, but it does correctly implement
			// parsing of the various Go strings.
			var sc scanner.Scanner
			fset := &token.FileSet{}
			file := fset.AddFile("", 0, len(args))
			sc.Init(file, []byte(args), nil, 0)
			_, tok, lit := sc.Scan()
			if tok != token.STRING || sc.ErrorCount != 0 {
				// Calculate start of token
				return nil, types.Error{
					Fset: p.program.fset,
					Pos:  patternPos,
					Msg:  "invalid quoted string in //go:embed",
				}
			}
			pattern := constant.StringVal(constant.MakeFromLiteral(lit, tok, 0))
			patterns = append(patterns, pattern)
			args = strings.TrimLeftFunc(args[len(lit):], unicode.IsSpace)
		default:
			// The value is just a regular value.
			// Split it at the first white space.
			index := strings.IndexFunc(args, unicode.IsSpace)
			if index < 0 {
				index = len(args)
			}
			pattern := args[:index]
			patterns = append(patterns, pattern)
			args = strings.TrimLeftFunc(args[len(pattern):], unicode.IsSpace)
		}
		if _, err := path.Match(patterns[len(patterns)-1], ""); err != nil {
			return nil, types.Error{
				Fset: p.program.fset,
				Pos:  patternPos,
				Msg:  "invalid pattern syntax",
			}
		}
	}
	return patterns, nil
}

// isValidEmbedType returns whether the given Go type can be used as a
// //go:embed type. This is only true for embed.FS, strings, and byte slices.
// The second return value indicates that this is a byte slice, and therefore
// the contents of the file needs to be passed to the compiler.
func isValidEmbedType(typ types.Type) (valid, byteSlice bool) {
	if typ.Underlying() == types.Typ[types.String] {
		// string type
		return true, false
	}
	if sliceType, ok := typ.Underlying().(*types.Slice); ok {
		if elemType, ok := sliceType.Elem().Underlying().(*types.Basic); ok && elemType.Kind() == types.Byte {
			// byte slice type
			return true, true
		}
	}
	if namedType, ok := typ.(*types.Named); ok && namedType.String() == "embed.FS" {
		// embed.FS type
		return true, false
	}
	return false, false
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

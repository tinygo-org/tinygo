package builder

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/interp"
	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

type syntheticSource struct {
	pkgPath string
	files   map[string]string
}

func makeLoaderPackages(t *testing.T, fset *token.FileSet, sources []syntheticSource) []*loader.Package {
	t.Helper()

	typeConf := types.Config{
		Importer: importer.Default(),
	}
	pkgs := make([]*loader.Package, 0, len(sources))
	for _, src := range sources {
		parsed := make([]*ast.File, 0, len(src.files))
		for filename, code := range src.files {
			f, err := parser.ParseFile(
				fset,
				filename,
				code,
				parser.ParseComments,
			)
			if err != nil {
				t.Fatal(err)
			}
			parsed = append(parsed, f)
		}
		info := types.Info{
			Types: map[ast.Expr]types.TypeAndValue{},
			Defs:  map[*ast.Ident]types.Object{},
			Uses:  map[*ast.Ident]types.Object{},
		}
		typesPkg, err := typeConf.Check(
			src.pkgPath,
			fset,
			parsed,
			&info,
		)
		if err != nil {
			t.Fatal(err)
		}
		pkgs = append(pkgs, &loader.Package{
			PackageJSON: loader.PackageJSON{ImportPath: src.pkgPath},
			Files:       parsed,
			Pkg:         typesPkg,
		})
	}
	return pkgs
}

func TestValidateWeakHooks(t *testing.T) {
	t.Parallel()

	t.Run("valid weak overrides and ignored non-weak decls", func(t *testing.T) {
		fset := token.NewFileSet()
		pkgs := makeLoaderPackages(t, fset, []syntheticSource{
			{
				pkgPath: "runtime",
				files: map[string]string{
					"runtime.go": `package runtime
import _ "unsafe"
//go:weak
func socInit0() {}
//go:linkname decomposeAny runtime.decomposeInterface
func decomposeAny(x any) (uintptr, uintptr)
func decomposeInterface(x struct{ a, b uintptr }) (uintptr, uintptr) { return x.a, x.b }
`,
				},
			},
			{
				pkgPath: "machine",
				files: map[string]string{
					"machine.go": `package machine
func boardXTAL() uint32
`,
				},
			},
			{
				pkgPath: "example.com/board",
				files: map[string]string{
					"board.go": `package board
import _ "unsafe"
//go:linkname mySocInit runtime.socInit0
func mySocInit() {}
func myXTAL() uint64 { return 12000000 }
//go:linkname myXTAL machine.boardXTAL
`,
				},
			},
			{
				pkgPath: "main",
				files: map[string]string{
					"main.go": `package main
func main() {}
`,
				},
			},
		})
		if err := validateWeakHooksFromPackages(pkgs, fset); err != nil {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("weak signature mismatch", func(t *testing.T) {
		fset := token.NewFileSet()
		pkgs := makeLoaderPackages(t, fset, []syntheticSource{
			{
				pkgPath: "runtime",
				files: map[string]string{
					"runtime.go": `package runtime
//go:weak
func socInit0() {}
`,
				},
			},
			{
				pkgPath: "example.com/board",
				files: map[string]string{
					"board.go": `package board
import _ "unsafe"
//go:linkname badInit runtime.socInit0
func badInit(x int) {}
`,
				},
			},
		})
		err := validateWeakHooksFromPackages(pkgs, fset)
		if err == nil {
			t.Fatal("expected signature mismatch error, got nil")
		}
		if !strings.Contains(err.Error(), "board.go:") || !strings.Contains(err.Error(), "does not match") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("method receiver linkname mismatch", func(t *testing.T) {
		fset := token.NewFileSet()
		pkgs := makeLoaderPackages(t, fset, []syntheticSource{
			{
				pkgPath: "runtime",
				files: map[string]string{
					"runtime.go": `package runtime
//go:weak
func socInit0() {}
`,
				},
			},
			{
				pkgPath: "example.com/board",
				files: map[string]string{
					"board.go": `package board
import _ "unsafe"
type T struct{ x int }
//go:linkname f runtime.socInit0
func (t T) f() {}
`,
				},
			},
		})
		err := validateWeakHooksFromPackages(pkgs, fset)
		if err == nil {
			t.Fatal("expected signature mismatch error for method receiver, got nil")
		}
		if !strings.Contains(err.Error(), "board.go:") || !strings.Contains(err.Error(), "does not match") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("file-level go:linknamestd weak mismatch", func(t *testing.T) {
		fset := token.NewFileSet()
		pkgs := makeLoaderPackages(t, fset, []syntheticSource{
			{
				pkgPath: "machine",
				files: map[string]string{
					"machine.go": `package machine
//go:weak
func boardXTAL() uint32 { return 0 }
`,
				},
			},
			{
				pkgPath: "example.com/board",
				files: map[string]string{
					"board.go": `package board
import _ "unsafe"
func badXTAL() uint64 { return 12000000 }
//go:linknamestd badXTAL machine.boardXTAL
`,
				},
			},
		})
		err := validateWeakHooksFromPackages(pkgs, fset)
		if err == nil {
			t.Fatal("expected signature mismatch error, got nil")
		}
		if !strings.Contains(err.Error(), "board.go:") || !strings.Contains(err.Error(), "does not match") {
			t.Errorf("unexpected error: %v", err)
		}
	})

	t.Run("duplicate weak definitions", func(t *testing.T) {
		fset := token.NewFileSet()
		pkgs := makeLoaderPackages(t, fset, []syntheticSource{
			{
				pkgPath: "runtime",
				files: map[string]string{
					"runtime.go": `package runtime
//go:weak
func socInit0() {}
`,
				},
			},
			{
				pkgPath: "machine",
				files: map[string]string{
					"machine.go": `package machine
import _ "unsafe"
//go:weak
//go:linkname dupWeak runtime.socInit0
func dupWeak() {}
`,
				},
			},
		})
		err := validateWeakHooksFromPackages(pkgs, fset)
		if err == nil || !strings.Contains(err.Error(), "duplicate //go:weak definition") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("duplicate strong overrides", func(t *testing.T) {
		fset := token.NewFileSet()
		pkgs := makeLoaderPackages(t, fset, []syntheticSource{
			{
				pkgPath: "runtime",
				files: map[string]string{
					"runtime.go": `package runtime
//go:weak
func boardInit0() {}
`,
				},
			},
			{
				pkgPath: "example.com/board1",
				files: map[string]string{
					"b1.go": `package board1
import _ "unsafe"
//go:linkname init1 runtime.boardInit0
func init1() {}
`,
				},
			},
			{
				pkgPath: "example.com/board2",
				files: map[string]string{
					"b2.go": `package board2
import _ "unsafe"
//go:linkname init2 runtime.boardInit0
func init2() {}
`,
				},
			},
		})
		err := validateWeakHooksFromPackages(pkgs, fset)
		if err == nil || !strings.Contains(err.Error(), "duplicate strong override for weak function") {
			t.Fatalf("unexpected error: %v", err)
		}
	})

	t.Run("linkname without unsafe is ignored", func(t *testing.T) {
		fset := token.NewFileSet()
		pkgs := makeLoaderPackages(t, fset, []syntheticSource{
			{
				pkgPath: "runtime",
				files: map[string]string{
					"runtime.go": `package runtime
//go:weak
func socInit0() {}
`,
				},
			},
			{
				pkgPath: "example.com/board",
				files: map[string]string{
					"board.go": `package board
//go:linkname notReally runtime.socInit0
func notReally(x int) {}
`,
				},
			},
		})
		if err := validateWeakHooksFromPackages(pkgs, fset); err != nil {
			t.Fatalf("unexpected error when unsafe is not imported: %v", err)
		}
	})
}

func makeWeakProbeTinyGoRoot(t *testing.T) string {
	t.Helper()

	origRoot := goenv.Get("TINYGOROOT")
	root := t.TempDir()

	entries, err := os.ReadDir(origRoot)
	if err != nil {
		t.Fatal(err)
	}
	for _, entry := range entries {
		if entry.Name() == "src" {
			continue
		}
		if err := os.Symlink(filepath.Join(origRoot, entry.Name()), filepath.Join(root, entry.Name())); err != nil {
			t.Fatal(err)
		}
	}

	// Mirror src directories cleanly so GetCachedGoroot sees real directories, not symlinks.
	origSrc := filepath.Join(origRoot, "src")
	dstSrc := filepath.Join(root, "src")
	err = filepath.WalkDir(origSrc, func(path string, d os.DirEntry, err error) error {
		if err != nil {
			return err
		}
		rel, err := filepath.Rel(origSrc, path)
		if err != nil {
			return err
		}
		targetPath := filepath.Join(dstSrc, rel)
		if d.IsDir() {
			return os.MkdirAll(targetPath, 0o755)
		}
		return os.Symlink(path, targetPath)
	})
	if err != nil {
		t.Fatal(err)
	}

	probeSrc := `package runtime

//go:weak
func weakProbeHook() int32 {
	return 7
}

var WeakProbeInitResult = weakProbeHook()
`
	if err := os.WriteFile(filepath.Join(dstSrc, "runtime", "weak_probe.go"), []byte(probeSrc), 0o600); err != nil {
		t.Fatal(err)
	}
	return root
}

func TestBuildWeakHookInitAndValidation(t *testing.T) {
	home := t.TempDir()
	t.Setenv("HOME", home)
	t.Setenv("XDG_CACHE_HOME", home)
	t.Setenv("LocalAppData", home)
	t.Setenv("GOCACHE", t.TempDir())
	tinygoRoot := makeWeakProbeTinyGoRoot(t)
	t.Setenv("TINYGOROOT", tinygoRoot)

	t.Run("own-package init uses override", func(t *testing.T) {
		for _, opt := range []string{"0", "z"} {
			t.Run("opt="+opt, func(t *testing.T) {
				dir := t.TempDir()
				files := map[string]string{
					"go.mod": "module example.com/weakinit\n\ngo 1.25.0\n",
					"main.go": `package main

import (
	"runtime"
	_ "unsafe"
)

//go:linkname override runtime.weakProbeHook
func override() int32 { return 99 }

func main() {
	println(runtime.WeakProbeInitResult, override())
}
`,
				}
				for name, content := range files {
					if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
						t.Fatal(err)
					}
				}

				outLL := filepath.Join(t.TempDir(), "out.ll")
				options := &compileopts.Options{
					Target:                  "wasm",
					Opt:                     opt,
					Directory:               dir,
					Semaphore:               make(chan struct{}, 2),
					InterpTimeout:           60 * time.Second,
					InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
				}
				config, err := NewConfig(options)
				if err != nil {
					t.Fatal(err)
				}
				if _, err := Build(".", outLL, t.TempDir(), config); err != nil {
					t.Fatalf("Build failed: %v", err)
				}

				ctx := llvm.NewContext()
				defer ctx.Dispose()
				buf, err := llvm.NewMemoryBufferFromFile(outLL)
				if err != nil {
					t.Fatal(err)
				}
				mod, err := ctx.ParseIR(buf)
				if err != nil {
					buf.Dispose()
					t.Fatal(err)
				}
				defer mod.Dispose()

				g := mod.NamedGlobal("runtime.WeakProbeInitResult")
				if !g.IsNil() {
					initVal := g.Initializer()
					if initVal.IsNil() || initVal.SExtValue() != 99 {
						t.Errorf("runtime.WeakProbeInitResult initializer = %v, want 99", initVal)
					}
				}
				if ir := mod.String(); strings.Contains(ir, "i32 7") {
					// Check main/init functions specifically in case unrelated code uses i32 7.
					for _, fnName := range []string{"main.main", "runtime.initAll", "runtime.weakProbeHook"} {
						fn := mod.NamedFunction(fnName)
						if !fn.IsNil() && strings.Contains(fn.String(), "i32 7") {
							t.Errorf("%s unexpectedly retained weak constant 7:\n%s", fnName, fn.String())
						}
					}
				}
			})
		}
	})

	t.Run("signature mismatch", func(t *testing.T) {
		dir := t.TempDir()
		files := map[string]string{
			"go.mod": "module example.com/hookcheck\n\ngo 1.25.0\n",
			"main.go": `package main

import _ "unsafe"

//go:linkname bad runtime.weakProbeHook
func bad() {}

func main() {}
`,
		}
		for name, content := range files {
			if err := os.WriteFile(filepath.Join(dir, name), []byte(content), 0o600); err != nil {
				t.Fatal(err)
			}
		}

		options := &compileopts.Options{
			Target:                  "wasm",
			Opt:                     "z",
			Directory:               dir,
			Semaphore:               make(chan struct{}, 2),
			InterpTimeout:           60 * time.Second,
			InterpMaxLoopIterations: interp.DefaultMaxInterpBlockEntries,
		}
		config, err := NewConfig(options)
		if err != nil {
			t.Fatal(err)
		}
		_, err = Build(
			".",
			filepath.Join(t.TempDir(), "out.wasm"),
			t.TempDir(),
			config,
		)
		if err == nil {
			t.Fatal("expected signature mismatch error from Build, got nil")
		}
		if !strings.Contains(err.Error(), "runtime.weakProbeHook") || !strings.Contains(err.Error(), "does not match") {
			t.Errorf("unexpected Build error: %v", err)
		}
	})
}

func TestWeakHookEndToEndDCE(t *testing.T) {
	t.Parallel()

	options := &compileopts.Options{Target: "wasm"}
	target, err := compileopts.LoadTarget(options)
	if err != nil {
		t.Fatal(err)
	}
	config := &compileopts.Config{Options: options, Target: target}
	compilerConfig := &compiler.Config{
		Triple:          config.Triple(),
		Features:        config.Features(),
		ABI:             config.ABI(),
		GOOS:            config.GOOS(),
		GOARCH:          config.GOARCH(),
		CodeModel:       config.CodeModel(),
		RelocationModel: config.RelocationModel(),
		Scheduler:       config.Scheduler(),
		PanicUnwind:     config.PanicUnwind(),
	}
	machine, err := compiler.NewTargetMachine(compilerConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer machine.Dispose()

	lprogram, err := loader.Load(config, "../compiler/testdata/basic.go", types.Config{Sizes: compiler.Sizes(machine)})
	if err != nil {
		t.Fatal(err)
	}
	templatePkg := *lprogram.Packages["runtime"]
	templatePkg.EmbedGlobals = nil

	compileToBC := func(pkgPath, filename, src string) string {
		fset := token.NewFileSet()
		f, err := parser.ParseFile(
			fset,
			filename,
			src,
			parser.ParseComments,
		)
		if err != nil {
			t.Fatal(err)
		}
		typeConf := types.Config{Importer: importer.Default(), Sizes: compiler.Sizes(machine)}
		info := types.Info{
			Types:      map[ast.Expr]types.TypeAndValue{},
			Defs:       map[*ast.Ident]types.Object{},
			Uses:       map[*ast.Ident]types.Object{},
			Selections: map[*ast.SelectorExpr]*types.Selection{},
		}
		typesPkg, err := typeConf.Check(
			pkgPath,
			fset,
			[]*ast.File{f},
			&info,
		)
		if err != nil {
			t.Fatal(err)
		}
		ssaProg := ssa.NewProgram(fset, ssa.BareInits|ssa.GlobalDebug|ssa.InstantiateGenerics)
		if pkgPath != "runtime" {
			rtFile, err := parser.ParseFile(
				fset,
				"rt.go",
				"package runtime\n",
				0,
			)
			if err != nil {
				t.Fatal(err)
			}
			rtInfo := types.Info{}
			rtPkg, err := typeConf.Check(
				"runtime",
				fset,
				[]*ast.File{rtFile},
				&rtInfo,
			)
			if err != nil {
				t.Fatal(err)
			}
			ssaProg.CreatePackage(
				rtPkg,
				[]*ast.File{rtFile},
				&rtInfo,
				true,
			)
		}
		ssaPkg := ssaProg.CreatePackage(
			typesPkg,
			[]*ast.File{f},
			&info,
			true,
		)
		ssaProg.Build()

		pkg := templatePkg
		pkg.Dir = t.TempDir()
		pkg.Pkg = typesPkg
		pkg.Files = []*ast.File{f}

		mod, errs := compiler.CompilePackage(
			pkgPath,
			&pkg,
			ssaPkg,
			machine,
			compilerConfig,
			false,
		)
		defer mod.Context().Dispose()
		defer mod.Dispose()
		if len(errs) != 0 {
			t.Fatalf("compile %s: %v", pkgPath, errs)
		}
		buf := llvm.WriteBitcodeToMemoryBuffer(mod)
		defer buf.Dispose()
		bcPath := filepath.Join(t.TempDir(), filepath.Base(filename)+".bc")
		if err := os.WriteFile(bcPath, buf.Bytes(), 0o600); err != nil {
			t.Fatal(err)
		}
		return bcPath
	}

	rtBC := compileToBC("runtime", "runtime_hook.go", `package runtime

//go:weak
func weakHook() int32 {
	return 7
}

//export useHook
func useHook() int32 {
	return weakHook()
}
`)
	boardBC := compileToBC("example.com/board", "board_hook.go", `package board

import _ "unsafe"

//go:linkname overrideHook runtime.weakHook
func overrideHook() int32 {
	return 99
}
`)

	for _, tc := range []struct {
		name    string
		bcs     []string
		wantRet int64
	}{
		{name: "unoverridden weak dce", bcs: []string{rtBC}, wantRet: 7},
		{name: "weak then strong", bcs: []string{rtBC, boardBC}, wantRet: 99},
		{name: "strong then weak", bcs: []string{boardBC, rtBC}, wantRet: 99},
	} {
		t.Run(tc.name, func(t *testing.T) {
			ctx := llvm.NewContext()
			defer ctx.Dispose()
			mod := ctx.NewModule("main")
			defer mod.Dispose()
			linker := llvm.NewLinker(mod)
			defer linker.Dispose()

			counts := map[string]linkSymbolCounts{}
			for _, bc := range tc.bcs {
				pkgMod, err := ctx.ParseBitcodeFile(bc)
				if err != nil {
					t.Fatal(err)
				}
				if err := linkPackageModule(
					linker,
					mod,
					pkgMod,
					counts,
				); err != nil {
					t.Fatal(err)
				}
			}

			internalizeModule(mod)
			hookFn := mod.NamedFunction("runtime.weakHook")
			if hookFn.IsNil() {
				t.Fatal("missing runtime.weakHook before optimization")
			}
			if hookFn.Linkage() != llvm.InternalLinkage {
				t.Errorf("hook linkage after internalization = %v, want InternalLinkage", hookFn.Linkage())
			}
			if hookFn.Visibility() != llvm.DefaultVisibility {
				t.Errorf("hook visibility after internalization = %v, want DefaultVisibility", hookFn.Visibility())
			}

			passOpts := llvm.NewPassBuilderOptions()
			defer passOpts.Dispose()
			if err := mod.RunPasses("default<O2>", llvm.TargetMachine{}, passOpts); err != nil {
				t.Fatal(err)
			}
			if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
				t.Fatal(err)
			}

			if !mod.NamedFunction("runtime.weakHook").IsNil() {
				t.Errorf("runtime.weakHook was not dead-stripped after O2:\n%s", mod.String())
			}
			useFn := mod.NamedFunction("useHook")
			if useFn.IsNil() {
				t.Fatal("missing useHook")
			}
			ret := useFn.FirstBasicBlock().FirstInstruction()
			if ret.IsAReturnInst().IsNil() || ret.Operand(0).SExtValue() != tc.wantRet {
				t.Errorf("useHook body = %s, want ret i32 %d", useFn.String(), tc.wantRet)
			}
		})
	}
}

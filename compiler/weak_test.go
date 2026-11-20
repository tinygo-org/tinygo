package compiler

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"slices"
	"strconv"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

func TestWeakPragmaErrors(t *testing.T) {
	t.Parallel()

	data, err := os.ReadFile("testdata/pragma-weak-errors.go")
	if err != nil {
		t.Fatal(err)
	}
	expected := []string{}
	for line := range strings.SplitSeq(string(data), "\n") {
		if after, ok := strings.CutPrefix(line, "// ERROR: "); ok {
			expected = append(expected, after)
		}
	}

	mod, errs := testCompilePackage(t, &compileopts.Options{Target: "wasm"}, "pragma-weak-errors.go")
	defer mod.Dispose()

	actual := []string{}
	for _, err := range errs {
		actual = append(actual, err.(types.Error).Msg)
	}
	slices.Sort(expected)
	slices.Sort(actual)
	if !slices.Equal(expected, actual) {
		t.Errorf("expected errors %q, got %q", expected, actual)
	}
}

func TestWeakPragmaRuntimeAndMachine(t *testing.T) {
	t.Parallel()

	t.Run("runtime weak", func(t *testing.T) {
		mod, errs := compileSyntheticPkg(
			t,
			"runtime",
			map[string]string{
				"hook.go": `package runtime

//go:weak
func socHook() int32 {
	return 7
}
`,
			},
			false,
		)
		defer mod.Dispose()
		if len(errs) != 0 {
			t.Fatalf("unexpected compile errors: %v", errs)
		}
		if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
			t.Fatal(err)
		}
		fn := mod.NamedFunction("runtime.socHook")
		if fn.IsNil() {
			t.Fatal("missing runtime.socHook")
		}
		if fn.Linkage() != llvm.WeakAnyLinkage {
			t.Errorf("linkage = %v, want WeakAnyLinkage", fn.Linkage())
		}
		if fn.Visibility() != llvm.HiddenVisibility {
			t.Errorf("visibility = %v, want HiddenVisibility", fn.Visibility())
		}
	})

	t.Run("machine weak", func(t *testing.T) {
		mod, errs := compileSyntheticPkg(
			t,
			"machine",
			map[string]string{
				"hook.go": `package machine

//go:weak
func boardHook() int32 {
	return 9
}
`,
			},
			false,
		)
		defer mod.Dispose()
		if len(errs) != 0 {
			t.Fatalf("unexpected compile errors: %v", errs)
		}
		if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
			t.Fatal(err)
		}
		fn := mod.NamedFunction("machine.boardHook")
		if fn.IsNil() {
			t.Fatal("missing machine.boardHook")
		}
		if fn.Linkage() != llvm.WeakAnyLinkage {
			t.Errorf("linkage = %v, want WeakAnyLinkage", fn.Linkage())
		}
		if fn.Visibility() != llvm.HiddenVisibility {
			t.Errorf("visibility = %v, want HiddenVisibility", fn.Visibility())
		}
	})

	t.Run("machine bodyless error", func(t *testing.T) {
		mod, errs := compileSyntheticPkg(
			t,
			"machine",
			map[string]string{
				"hook.go": `package machine

//go:weak
func missingBody()
`,
			},
			false,
		)
		defer mod.Dispose()
		assertSingleErrorMsg(t, errs, "can only use //go:weak on definitions")
	})

	t.Run("runtime args error", func(t *testing.T) {
		mod, errs := compileSyntheticPkg(
			t,
			"runtime",
			map[string]string{
				"hook.go": `package runtime

//go:weak bad
func withArg() {}
`,
			},
			false,
		)
		defer mod.Dispose()
		assertSingleErrorMsg(t, errs, "//go:weak does not accept parameters")
	})

	t.Run("file-level linknamestd", func(t *testing.T) {
		mod, errs := compileSyntheticPkg(
			t,
			"runtime",
			map[string]string{
				"hook.go": `package runtime

import _ "unsafe"

func localStd() int32 {
	return 5
}

//go:linknamestd localStd runtime.targetStd
`,
			},
			false,
		)
		defer mod.Dispose()
		if len(errs) != 0 {
			t.Fatalf("unexpected compile errors: %v", errs)
		}
		if mod.NamedFunction("runtime.targetStd").IsNil() {
			t.Fatal("missing runtime.targetStd for file-level //go:linknamestd")
		}
	})
}

func TestWeakSamePackageOverride(t *testing.T) {
	t.Parallel()

	weakSrc := `package runtime

//go:weak
func socInit0(x int32) int32 {
	if x > 0 {
		return 1
	}
	return 2
}
`
	strongSrc := `package runtime

import _ "unsafe"

//go:linkname rp2SocInit0 runtime.socInit0
func rp2SocInit0(x int32) int32 {
	return 42
}
`

	for _, debug := range []bool{false, true} {
		for _, tc := range []struct {
			name  string
			files map[string]string
		}{
			{
				name: "weak before strong",
				files: map[string]string{
					"a_weak.go":   weakSrc,
					"b_strong.go": strongSrc,
				},
			},
			{
				name: "strong before weak",
				files: map[string]string{
					"a_strong.go": strongSrc,
					"b_weak.go":   weakSrc,
				},
			},
		} {
			t.Run(tc.name+"/debug="+strconv.FormatBool(debug), func(t *testing.T) {
				mod, errs := compileSyntheticPkg(t, "runtime", tc.files, debug)
				defer mod.Dispose()
				if len(errs) != 0 {
					t.Fatalf("unexpected compile errors: %v", errs)
				}
				if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
					t.Fatal(err)
				}
				fn := mod.NamedFunction("runtime.socInit0")
				if fn.IsNil() {
					t.Fatal("missing runtime.socInit0")
				}
				if fn.Linkage() != llvm.ExternalLinkage {
					t.Errorf("linkage = %v, want ExternalLinkage", fn.Linkage())
				}
				if fn.Visibility() != llvm.HiddenVisibility {
					t.Errorf("visibility = %v, want HiddenVisibility", fn.Visibility())
				}
				if blocks := fn.BasicBlocksCount(); blocks != 1 {
					t.Errorf("basic blocks = %d, want 1 (strong body)", blocks)
				}
				ret := fn.FirstBasicBlock().LastInstruction()
				if ret.IsAReturnInst().IsNil() || ret.Operand(0).SExtValue() != 42 {
					t.Errorf("unexpected return instruction: %s", fn.String())
				}
			})
		}
	}

	errorCases := []struct {
		name    string
		wantErr string
		files   map[string]string
	}{
		{
			name:    "duplicate weak",
			wantErr: "redeclared in this program",
			files: map[string]string{
				"a.go": `package runtime
//go:weak
func dupWeak() {}
`,
				"b.go": `package runtime
import _ "unsafe"
//go:weak
//go:linkname dupWeak2 runtime.dupWeak
func dupWeak2() {}
`,
			},
		},
		{
			name:    "weak and two strongs",
			wantErr: "redeclared in this program",
			files: map[string]string{
				"a.go": `package runtime
//go:weak
func hook() {}
`,
				"b.go": `package runtime
import _ "unsafe"
//go:linkname s1 runtime.hook
func s1() {}
//go:linkname s2 runtime.hook
func s2() {}
`,
			},
		},
		{
			name:    "strong then weak then strong",
			wantErr: "redeclared in this program",
			files: map[string]string{
				"a.go": `package runtime
import _ "unsafe"
//go:linkname s1 runtime.hook
func s1() {}
`,
				"b.go": `package runtime
//go:weak
func hook() {}
`,
				"c.go": `package runtime
import _ "unsafe"
//go:linkname s2 runtime.hook
func s2() {}
`,
			},
		},
		{
			name:    "strong and two weaks",
			wantErr: "redeclared in this program",
			files: map[string]string{
				"a.go": `package runtime
import _ "unsafe"
//go:linkname s1 runtime.hook
func s1() {}
`,
				"b.go": `package runtime
//go:weak
func hook() {}
`,
				"c.go": `package runtime
import _ "unsafe"
//go:weak
//go:linkname w2 runtime.hook
func w2() {}
`,
			},
		},
		{
			name:    "signature mismatch weak then strong",
			wantErr: "does not match previous definition",
			files: map[string]string{
				"a.go": `package runtime
//go:weak
func hook() int32 { return 1 }
`,
				"b.go": `package runtime
import _ "unsafe"
//go:linkname badStrong runtime.hook
func badStrong(x int32) int32 { return x }
`,
			},
		},
		{
			name:    "signature mismatch strong then weak",
			wantErr: "does not match previous definition",
			files: map[string]string{
				"a.go": `package runtime
import _ "unsafe"
//go:linkname badStrong runtime.hook
func badStrong(x int32) int32 { return x }
`,
				"b.go": `package runtime
//go:weak
func hook() int32 { return 1 }
`,
			},
		},
	}
	for _, tc := range errorCases {
		t.Run(tc.name, func(t *testing.T) {
			mod, errs := compileSyntheticPkg(t, "runtime", tc.files, false)
			defer mod.Dispose()
			if len(errs) == 0 {
				t.Fatalf("expected error %q, got none", tc.wantErr)
			}
			if !strings.Contains(errs[0].Error(), tc.wantErr) {
				t.Errorf("error = %q, want substring %q", errs[0].Error(), tc.wantErr)
			}
		})
	}
}

func assertSingleErrorMsg(t *testing.T, errs []error, want string) {
	t.Helper()
	if len(errs) != 1 {
		t.Fatalf("got %d errors %v, want 1 (%q)", len(errs), errs, want)
	}
	if got := errs[0].Error(); !strings.Contains(got, want) {
		t.Errorf("error = %q, want substring %q", got, want)
	}
}

func compileSyntheticPkg(t *testing.T, pkgPath string, files map[string]string, debug bool) (llvm.Module, []error) {
	t.Helper()

	options := &compileopts.Options{Target: "wasm"}
	target, err := compileopts.LoadTarget(options)
	if err != nil {
		t.Fatal(err)
	}
	config := &compileopts.Config{Options: options, Target: target}
	compilerConfig := &Config{
		Triple:          config.Triple(),
		Features:        config.Features(),
		ABI:             config.ABI(),
		GOOS:            config.GOOS(),
		GOARCH:          config.GOARCH(),
		CodeModel:       config.CodeModel(),
		RelocationModel: config.RelocationModel(),
		Scheduler:       config.Scheduler(),
		PanicUnwind:     config.PanicUnwind(),
		Debug:           debug,
	}
	machine, err := NewTargetMachine(compilerConfig)
	if err != nil {
		t.Fatal(err)
	}
	defer machine.Dispose()

	lprogram, err := loader.Load(config, "./testdata/basic.go", types.Config{Sizes: Sizes(machine)})
	if err != nil {
		t.Fatal(err)
	}
	templatePkg := *lprogram.Packages["runtime"]
	templatePkg.Dir = t.TempDir()
	templatePkg.EmbedGlobals = nil

	fset := token.NewFileSet()
	names := make([]string, 0, len(files))
	for name := range files {
		names = append(names, name)
	}
	slices.Sort(names)

	parsed := make([]*ast.File, 0, len(names))
	for _, name := range names {
		f, err := parser.ParseFile(
			fset,
			filepath.Join(templatePkg.Dir, name),
			files[name],
			parser.ParseComments,
		)
		if err != nil {
			t.Fatal(err)
		}
		parsed = append(parsed, f)
	}

	typeConf := types.Config{
		Importer: importer.Default(),
		Sizes:    Sizes(machine),
	}
	info := types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}
	typesPkg, err := typeConf.Check(
		pkgPath,
		fset,
		parsed,
		&info,
	)
	if err != nil {
		t.Fatal(err)
	}

	ssaProg := ssa.NewProgram(fset, ssa.BareInits|ssa.GlobalDebug|ssa.InstantiateGenerics)
	if pkgPath != "runtime" {
		rtFile, err := parser.ParseFile(
			fset,
			filepath.Join(templatePkg.Dir, "rt.go"),
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
		parsed,
		&info,
		true,
	)
	ssaProg.Build()

	templatePkg.Pkg = typesPkg
	templatePkg.Files = parsed
	return CompilePackage(
		pkgPath,
		&templatePkg,
		ssaPkg,
		machine,
		compilerConfig,
		false,
	)
}

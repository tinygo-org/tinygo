package compiler

import (
	"go/ast"
	"go/importer"
	"go/parser"
	"go/token"
	"go/types"
	"path/filepath"
	"slices"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/loader"
	"golang.org/x/tools/go/ssa"
	"tinygo.org/x/go-llvm"
)

func TestNoHeapInitHooks(t *testing.T) {
	t.Parallel()

	mod, errs := compileNoHeapTestPkg(
		t,
		"example.com/board",
		map[string]string{
			"hooks.go": `package board

import _ "unsafe"

//go:linkname myBoardInit0 runtime.boardInit0
func myBoardInit0() *int32 {
	return new(int32)
}

//go:linkname mySocInit0 runtime.socInit0
func mySocInit0() *int32 {
	return new(int32)
}

//go:linkname myHwInit0 runtime.hwInit0
func myHwInit0() *int32 {
	return new(int32)
}

//go:linkname myBoardInit1 runtime.boardInit1
func myBoardInit1() *int32 {
	return new(int32)
}

//go:linkname mySocInit1 runtime.socInit1
func mySocInit1() *int32 {
	return new(int32)
}

//go:linkname myHwInit1 runtime.hwInit1
func myHwInit1() *int32 {
	return new(int32)
}
`,
		},
	)
	defer mod.Dispose()
	if len(errs) != 0 {
		t.Fatalf("unexpected compile errors: %v", errs)
	}
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}

	for _, name := range []string{"runtime.hwInit0", "runtime.socInit0", "runtime.boardInit0"} {
		ir := mod.NamedFunction(name).String()
		if !strings.Contains(ir, "@runtime.alloc_noheap") {
			t.Errorf("%s did not use runtime.alloc_noheap:\n%s", name, ir)
		}
		if strings.Contains(ir, "@runtime.alloc(") {
			t.Errorf("%s unexpectedly used runtime.alloc:\n%s", name, ir)
		}
	}
	for _, name := range []string{"runtime.hwInit1", "runtime.socInit1", "runtime.boardInit1"} {
		ir := mod.NamedFunction(name).String()
		if !strings.Contains(ir, "@runtime.alloc(") {
			t.Errorf("%s did not use runtime.alloc:\n%s", name, ir)
		}
		if strings.Contains(ir, "@runtime.alloc_noheap") {
			t.Errorf("%s unexpectedly used runtime.alloc_noheap:\n%s", name, ir)
		}
	}
}

func compileNoHeapTestPkg(t *testing.T, pkgPath string, files map[string]string) (llvm.Module, []error) {
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
	rtFile, err := parser.ParseFile(
		fset,
		filepath.Join(templatePkg.Dir, "rt.go"),
		`package runtime
import "unsafe"
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer
func alloc_noheap(size uintptr, layout unsafe.Pointer) unsafe.Pointer
func trackPointer(ptr, alloca unsafe.Pointer)
`,
		0,
	)
	if err != nil {
		t.Fatal(err)
	}
	rtInfo := types.Info{
		Types:      map[ast.Expr]types.TypeAndValue{},
		Defs:       map[*ast.Ident]types.Object{},
		Uses:       map[*ast.Ident]types.Object{},
		Selections: map[*ast.SelectorExpr]*types.Selection{},
	}
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

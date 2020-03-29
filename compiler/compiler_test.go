package compiler

import (
	"bytes"
	"flag"
	"fmt"
	"go/ast"
	"go/parser"
	"go/token"
	"go/types"
	"io/ioutil"
	"path/filepath"
	"strings"
	"sync"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler/ircheck"
	"golang.org/x/tools/go/ssa"
	"golang.org/x/tools/go/ssa/ssautil"
	"tinygo.org/x/go-llvm"
)

var flagUpdate = flag.Bool("update", false, "update all tests")

func TestCompiler(t *testing.T) {
	t.Parallel()
	for _, name := range []string{"basic"} {
		t.Run(name, func(t *testing.T) {
			runCompilerTest(t, name)
		})
	}
}

func runCompilerTest(t *testing.T, name string) {
	// Read the AST in memory.
	path := filepath.Join("testdata", name+".go")
	fset := token.NewFileSet()
	f, err := parser.ParseFile(fset, path, nil, parser.ParseComments)
	if err != nil {
		t.Fatal("could not parse Go source file:", err)
	}
	files := []*ast.File{f}

	// Create Go SSA from the AST.
	var typecheckErrors []error
	var typecheckErrorsLock sync.Mutex
	typesConfig := types.Config{
		Error: func(err error) {
			typecheckErrorsLock.Lock()
			defer typecheckErrorsLock.Unlock()
			typecheckErrors = append(typecheckErrors, err)
		},
		Importer: simpleImporter{},
		Sizes:    types.SizesFor("gccgo", "arm"),
	}
	pkg, _, err := ssautil.BuildPackage(&typesConfig, fset, types.NewPackage("main", ""), files, ssa.SanityCheckFunctions|ssa.BareInits|ssa.GlobalDebug)
	for _, err := range typecheckErrors {
		t.Error(err)
	}
	if err != nil && len(typecheckErrors) == 0 {
		// Only report errors when no type errors are found (an
		// unexpected condition).
		t.Error(err)
	}
	if t.Failed() {
		return
	}

	// Configure the compiler.
	config := compileopts.Config{
		Options: &compileopts.Options{},
		Target: &compileopts.TargetSpec{
			Triple:    "armv7m-none-eabi",
			BuildTags: []string{"cortexm", "baremetal", "linux", "arm"},
			Scheduler: "tasks",
		},
	}
	machine, err := NewTargetMachine(&config)
	if err != nil {
		t.Fatal(err)
	}
	c := newCompilerContext("main", machine, &config)
	irbuilder := c.ctx.NewBuilder()
	defer irbuilder.Dispose()

	// Create LLVM IR from the Go SSA.
	c.createPackage(pkg, irbuilder)

	// Check the IR with the LLVM verifier.
	if err := llvm.VerifyModule(c.mod, llvm.PrintMessageAction); err != nil {
		t.Error("verification error after IR construction")
	}

	// Check the IR with our own verifier (which checks for different things).
	errs := ircheck.Module(c.mod)
	for _, err := range errs {
		t.Error(err)
	}

	// Check whether the IR matches the expected IR.
	ir := c.mod.String()
	ir = ir[strings.Index(ir, "\ntarget datalayout = ")+1:]
	outfile := filepath.Join("testdata", name+".ll")
	if *flagUpdate {
		err := ioutil.WriteFile(outfile, []byte(ir), 0666)
		if err != nil {
			t.Error("could not read output file:", err)
		}
	} else {
		ir2, err := ioutil.ReadFile(outfile)
		if err != nil {
			t.Fatal("could not read input file:", err)
		}
		ir2 = bytes.Replace(ir2, []byte("\r\n"), []byte("\n"), -1)
		if ir != string(ir2) {
			t.Error("output did not match")
		}
	}
}

// simpleImporter implements the types.Importer interface, but only allows
// importing the unsafe package.
type simpleImporter struct {
}

// Import implements the Importer interface. For testing usage only: it only
// supports importing the unsafe package.
func (i simpleImporter) Import(path string) (*types.Package, error) {
	switch path {
	case "unsafe":
		return types.Unsafe, nil
	default:
		return nil, fmt.Errorf("importer not implemented for package %s", path)
	}
}

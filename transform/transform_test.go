package transform_test

// This file defines some helper functions for testing transforms.

import (
	"flag"
	"go/types"
	"os"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/internal/ir"
	"github.com/tinygo-org/tinygo/loader"
	"tinygo.org/x/go-llvm"
)

var update = flag.Bool("update", false, "update transform package tests")

var defaultTestConfig = &compileopts.Config{
	Target:  &compileopts.TargetSpec{},
	Options: &compileopts.Options{Opt: "2"},
}

// checkGolden compares got against the golden file at path, or rewrites that
// file when the test is run with -update.
func checkGolden(t *testing.T, path, got string) {
	if *update {
		if err := os.WriteFile(path, []byte(got), 0o666); err != nil {
			t.Fatal(err)
		}
		return
	}
	want, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	if got != string(want) {
		t.Errorf("%s does not match expected output:\n%s", path, got)
	}
}

// testTransform runs a transformation pass on an input file (pathPrefix+".ll")
// and checks whether it matches the expected output (pathPrefix+".out.ll"). The
// output is compared with a fuzzy match that ignores some irrelevant lines such
// as empty lines.
func testTransform(t *testing.T, pathPrefix string, transform func(mod llvm.Module)) {
	// Read the input IR.
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile(pathPrefix + ".ll")
	if err != nil {
		t.Fatalf("could not read file %s: %v", pathPrefix+".ll", err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatalf("could not load module:\n%v", err)
	}
	defer mod.Dispose()

	// Perform the transform.
	transform(mod)

	// Check for any incorrect IR.
	err = llvm.VerifyModule(mod, llvm.PrintMessageAction)
	if err != nil {
		t.Fatal("IR verification failed")
	}

	// Get the output from the test and filter some irrelevant lines.
	actual := mod.String()
	actual = actual[strings.Index(actual, "\ntarget datalayout = ")+1:]

	if *update {
		err := os.WriteFile(pathPrefix+".out.ll", []byte(ir.Normalize(actual)), 0666)
		if err != nil {
			t.Error("failed to write out new output:", err)
		}
	} else {
		// Read the expected output IR.
		out, err := os.ReadFile(pathPrefix + ".out.ll")
		if err != nil {
			t.Fatalf("could not read output file %s: %v", pathPrefix+".out.ll", err)
		}

		// See whether the transform output matches with the expected output IR.
		expected := string(out)
		if diff := ir.Diff(expected, actual); diff != "" {
			t.Errorf("output does not match expected output (re-run with -update to regenerate):\n%s", diff)
		}
	}
}

// compileGoFileForTesting compiles the given Go file to run tests against.
// Only the given Go file is compiled (no dependencies) and no optimizations are
// run.
// If there are any errors, they are reported via the *testing.T instance.
func compileGoFileForTesting(t *testing.T, filename string) llvm.Module {
	target, err := compileopts.LoadTarget(&compileopts.Options{GOOS: "linux", GOARCH: "386"})
	if err != nil {
		t.Fatal("failed to load target:", err)
	}
	config := &compileopts.Config{
		Options: &compileopts.Options{},
		Target:  target,
	}
	compilerConfig := &compiler.Config{
		Triple:             config.Triple(),
		GOOS:               config.GOOS(),
		GOARCH:             config.GOARCH(),
		CodeModel:          config.CodeModel(),
		RelocationModel:    config.RelocationModel(),
		Scheduler:          config.Scheduler(),
		AutomaticStackSize: config.AutomaticStackSize(),
		Debug:              true,
		PanicStrategy:      config.PanicStrategy(),
		PanicUnwind:        config.PanicUnwind(),
	}
	machine, err := compiler.NewTargetMachine(compilerConfig)
	if err != nil {
		t.Fatal("failed to create target machine:", err)
	}
	defer machine.Dispose()

	// Load entire program AST into memory.
	lprogram, err := loader.Load(config, filename, types.Config{
		Sizes: compiler.Sizes(machine),
	})
	if err != nil {
		t.Fatal("failed to create target machine:", err)
	}
	err = lprogram.Parse()
	if err != nil {
		t.Fatal("could not parse", err)
	}

	// Compile AST to IR.
	program := lprogram.LoadSSA()
	pkg := lprogram.MainPkg()
	ssaPkg := program.Package(pkg.Pkg)
	ssaPkg.Build()
	mod, errs := compiler.CompilePackage(filename, pkg, ssaPkg, machine, compilerConfig, false)
	if errs != nil {
		for _, err := range errs {
			t.Error(err)
		}
		t.FailNow()
	}
	return mod
}

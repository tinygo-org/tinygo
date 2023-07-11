package transform_test

// This file defines some helper functions for testing transforms.

import (
	"flag"
	"go/token"
	"go/types"
	"os"
	"path/filepath"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/compiler"
	"github.com/tinygo-org/tinygo/loader"
	"tinygo.org/x/go-llvm"
)

var update = flag.Bool("update", false, "update transform package tests")

var defaultTestConfig = &compileopts.Config{
	Target:  &compileopts.TargetSpec{},
	Options: &compileopts.Options{Opt: "2"},
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
	os.Stat(pathPrefix + ".ll") // make sure this file is tracked by `go test` caching
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
		err := os.WriteFile(pathPrefix+".out.ll", []byte(actual), 0666)
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
		if !fuzzyEqualIR(expected, actual) {
			t.Logf("output does not match expected output:\n%s", actual)
			t.Fail()
		}
	}
}

// fuzzyEqualIR returns true if the two LLVM IR strings passed in are roughly
// equal. That means, only relevant lines are compared (excluding comments
// etc.).
func fuzzyEqualIR(s1, s2 string) bool {
	lines1 := filterIrrelevantIRLines(strings.Split(s1, "\n"))
	lines2 := filterIrrelevantIRLines(strings.Split(s2, "\n"))
	if len(lines1) != len(lines2) {
		return false
	}
	for i, line1 := range lines1 {
		line2 := lines2[i]
		if line1 != line2 {
			return false
		}
	}

	return true
}

// filterIrrelevantIRLines removes lines from the input slice of strings that
// are not relevant in comparing IR. For example, empty lines and comments are
// stripped out.
func filterIrrelevantIRLines(lines []string) []string {
	var out []string
	for _, line := range lines {
		line = strings.Split(line, ";")[0]    // strip out comments/info
		line = strings.TrimRight(line, "\r ") // drop '\r' on Windows and remove trailing spaces from comments
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "source_filename = ") {
			continue
		}
		out = append(out, line)
	}
	return out
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
	}
	machine, err := compiler.NewTargetMachine(compilerConfig)
	if err != nil {
		t.Fatal("failed to create target machine:", err)
	}
	defer machine.Dispose()

	// Load entire program AST into memory.
	lprogram, err := loader.Load(config, filename, config.ClangHeaders, types.Config{
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
	mod, errs := compiler.CompilePackage(filename, pkg, program.Package(pkg.Pkg), machine, compilerConfig, false)
	if errs != nil {
		for _, err := range errs {
			t.Error(err)
		}
		t.FailNow()
	}
	return mod
}

// getPosition returns the position information for the given value, as far as
// it is available.
func getPosition(val llvm.Value) token.Position {
	if !val.IsAInstruction().IsNil() {
		loc := val.InstructionDebugLoc()
		if loc.IsNil() {
			return token.Position{}
		}
		file := loc.LocationScope().ScopeFile()
		return token.Position{
			Filename: filepath.Join(file.FileDirectory(), file.FileFilename()),
			Line:     int(loc.LocationLine()),
			Column:   int(loc.LocationColumn()),
		}
	} else if !val.IsAFunction().IsNil() {
		loc := val.Subprogram()
		if loc.IsNil() {
			return token.Position{}
		}
		file := loc.ScopeFile()
		return token.Position{
			Filename: filepath.Join(file.FileDirectory(), file.FileFilename()),
			Line:     int(loc.SubprogramLine()),
		}
	} else {
		return token.Position{}
	}
}

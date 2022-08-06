package compiler

import (
	"flag"
	"go/types"
	"os"
	"strconv"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/goenv"
	"github.com/tinygo-org/tinygo/loader"
	"tinygo.org/x/go-llvm"
)

// Pass -update to go test to update the output of the test files.
var flagUpdate = flag.Bool("update", false, "update tests based on test output")

type testCase struct {
	file      string
	target    string
	scheduler string
}

// Basic tests for the compiler. Build some Go files and compare the output with
// the expected LLVM IR for regression testing.
func TestCompiler(t *testing.T) {
	t.Parallel()

	// Determine Go minor version (e.g. 16 in go1.16.3).
	_, goMinor, err := goenv.GetGorootVersion()
	if err != nil {
		t.Fatal("could not read Go version:", err)
	}

	// Determine which tests to run, depending on the Go and LLVM versions.
	tests := []testCase{
		{"basic.go", "", ""},
		{"pointer.go", "", ""},
		{"slice.go", "", ""},
		{"string.go", "", ""},
		{"float.go", "", ""},
		{"interface.go", "", ""},
		{"func.go", "", ""},
		{"defer.go", "cortex-m-qemu", ""},
		{"pragma.go", "", ""},
		{"goroutine.go", "wasm", "asyncify"},
		{"goroutine.go", "cortex-m-qemu", "tasks"},
		{"channel.go", "", ""},
		{"gc.go", "", ""},
		{"goasm.go", "linux/amd64", ""},
		{"goasm.go", "linux/arm64", ""},
		{"zeromap.go", "", ""},
	}

	// For the goasm.go test.
	goAsmReferences := map[string]string{
		"main.AsmSqrt":          "__GoABI0_main.AsmSqrt",
		"main.AsmAdd":           "__GoABI0_main.AsmAdd",
		"main.AsmFoo":           "__GoABI0_main.AsmFoo",
		"main.AsmAllIntTypes":   "__GoABI0_main.AsmAllIntTypes",
		"main.AsmAllFloatTypes": "__GoABI0_main.AsmAllFloatTypes",
		"main.AsmAllOtherTypes": "__GoABI0_main.AsmAllOtherTypes",
		"main.asmExport":        "__GoABI0_main.asmExport",
		"main.asmGlobalExport":  "__GoABI0_main.asmGlobalExport",
	}
	if goMinor >= 20 {
		tests = append(tests, testCase{"go1.20.go", "", ""})
	}
	if goMinor >= 21 {
		tests = append(tests, testCase{"go1.21.go", "", ""})
	}

	for _, tc := range tests {
		name := tc.file
		targetString := "wasm"
		if tc.target != "" {
			targetString = tc.target
			name += "-" + tc.target
		}
		if tc.scheduler != "" {
			name += "-" + tc.scheduler
		}

		t.Run(name, func(t *testing.T) {
			options := &compileopts.Options{}
			if strings.Contains(targetString, "/") {
				parts := strings.Split(targetString, "/")
				options.GOOS = parts[0]
				options.GOARCH = parts[1]
			} else {
				options.Target = targetString
			}
			if tc.scheduler != "" {
				options.Scheduler = tc.scheduler
			}

			mod, errs := testCompilePackage(t, options, tc.file, goAsmReferences)
			if errs != nil {
				for _, err := range errs {
					t.Error(err)
				}
				return
			}

			err := llvm.VerifyModule(mod, llvm.PrintMessageAction)
			if err != nil {
				t.Error(err)
			}

			// Optimize IR a little.
			passOptions := llvm.NewPassBuilderOptions()
			defer passOptions.Dispose()
			err = mod.RunPasses("instcombine", llvm.TargetMachine{}, passOptions)
			if err != nil {
				t.Error(err)
			}

			outFilePrefix := tc.file[:len(tc.file)-3]
			if tc.target != "" {
				outFilePrefix += "-" + strings.ReplaceAll(tc.target, "/", "_")
			}
			if tc.scheduler != "" {
				outFilePrefix += "-" + tc.scheduler
			}
			outPath := "./testdata/" + outFilePrefix + ".ll"

			// Update test if needed. Do not check the result.
			if *flagUpdate {
				err := os.WriteFile(outPath, []byte(mod.String()), 0666)
				if err != nil {
					t.Error("failed to write updated output file:", err)
				}
				return
			}

			expected, err := os.ReadFile(outPath)
			if err != nil {
				t.Fatal("failed to read golden file:", err)
			}

			if !fuzzyEqualIR(mod.String(), string(expected)) {
				t.Errorf("output does not match expected output:\n%s", mod.String())
			}
		})
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
	llvmVersion, err := strconv.Atoi(strings.Split(llvm.Version, ".")[0])
	if err != nil {
		// Note: this should never happen and if it does, it will always happen
		// for a particular build because llvm.Version is a constant.
		panic(err)
	}
	for _, line := range lines {
		line = strings.Split(line, ";")[0]    // strip out comments/info
		line = strings.TrimRight(line, "\r ") // drop '\r' on Windows and remove trailing spaces from comments
		if line == "" {
			continue
		}
		if strings.HasPrefix(line, "source_filename = ") {
			continue
		}
		if llvmVersion < 15 && strings.HasPrefix(line, "target datalayout = ") {
			// The datalayout string may vary betewen LLVM versions.
			// Right now test outputs are for LLVM 15 and higher.
			continue
		}
		out = append(out, line)
	}
	return out
}

func TestCompilerErrors(t *testing.T) {
	t.Parallel()

	// Read expected errors from the test file.
	var expectedErrors []string
	errorsFile, err := os.ReadFile("testdata/errors.go")
	if err != nil {
		t.Error(err)
	}
	errorsFileString := strings.ReplaceAll(string(errorsFile), "\r\n", "\n")
	for _, line := range strings.Split(errorsFileString, "\n") {
		if strings.HasPrefix(line, "// ERROR: ") {
			expectedErrors = append(expectedErrors, strings.TrimPrefix(line, "// ERROR: "))
		}
	}

	// Compile the Go file with errors.
	options := &compileopts.Options{
		Target: "wasm",
	}
	_, errs := testCompilePackage(t, options, "errors.go", nil)

	// Check whether the actual errors match the expected errors.
	expectedErrorsIdx := 0
	for _, err := range errs {
		err := err.(types.Error)
		position := err.Fset.Position(err.Pos)
		position.Filename = "errors.go" // don't use a full path
		if expectedErrorsIdx >= len(expectedErrors) || expectedErrors[expectedErrorsIdx] != err.Msg {
			t.Errorf("unexpected compiler error: %s: %s", position.String(), err.Msg)
			continue
		}
		expectedErrorsIdx++
	}
}

// Build a package given a number of compiler options and a file.
func testCompilePackage(t *testing.T, options *compileopts.Options, file string, goAsmReferences map[string]string) (llvm.Module, []error) {
	target, err := compileopts.LoadTarget(options)
	if err != nil {
		t.Fatal("failed to load target:", err)
	}
	config := &compileopts.Config{
		Options: options,
		Target:  target,
	}
	compilerConfig := &Config{
		Triple:             config.Triple(),
		Features:           config.Features(),
		ABI:                config.ABI(),
		GOOS:               config.GOOS(),
		GOARCH:             config.GOARCH(),
		CodeModel:          config.CodeModel(),
		RelocationModel:    config.RelocationModel(),
		Scheduler:          config.Scheduler(),
		AutomaticStackSize: config.AutomaticStackSize(),
		DefaultStackSize:   config.StackSize(),
		NeedsStackObjects:  config.NeedsStackObjects(),
	}
	machine, err := NewTargetMachine(compilerConfig)
	if err != nil {
		t.Fatal("failed to create target machine:", err)
	}
	defer machine.Dispose()

	// Load entire program AST into memory.
	lprogram, err := loader.Load(config, "./testdata/"+file, config.ClangHeaders, types.Config{
		Sizes: Sizes(machine),
	})
	if err != nil {
		t.Fatal("failed to create target machine:", err)
	}
	err = lprogram.Parse()
	if err != nil {
		t.Fatalf("could not parse test case %s: %s", file, err)
	}

	// Compile AST to IR.
	program := lprogram.LoadSSA()
	pkg := lprogram.MainPkg()
	return CompilePackage(file, pkg, program.Package(pkg.Pkg), machine, compilerConfig, goAsmReferences, false)
}

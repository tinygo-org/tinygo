package compiler

import (
	"flag"
	"go/types"
	"io/ioutil"
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
	file   string
	target string
}

// Basic tests for the compiler. Build some Go files and compare the output with
// the expected LLVM IR for regression testing.
func TestCompiler(t *testing.T) {
	t.Parallel()

	// Check LLVM version.
	llvmMajor, err := strconv.Atoi(strings.SplitN(llvm.Version, ".", 2)[0])
	if err != nil {
		t.Fatal("could not parse LLVM version:", llvm.Version)
	}
	if llvmMajor < 11 {
		// It is likely this version needs to be bumped in the future.
		// The goal is to at least test the LLVM version that's used by default
		// in TinyGo and (if possible without too many workarounds) also some
		// previous versions.
		t.Skip("compiler tests require LLVM 11 or above, got LLVM ", llvm.Version)
	}

	tests := []testCase{
		{"basic.go", ""},
		{"pointer.go", ""},
		{"slice.go", ""},
		{"string.go", ""},
		{"float.go", ""},
		{"interface.go", ""},
		{"func.go", ""},
		{"pragma.go", ""},
		{"goroutine.go", "wasm"},
		{"goroutine.go", "cortex-m-qemu"},
		{"channel.go", ""},
		{"intrinsics.go", "cortex-m-qemu"},
		{"intrinsics.go", "wasm"},
		{"gc.go", ""},
	}

	_, minor, err := goenv.GetGorootVersion(goenv.Get("GOROOT"))
	if err != nil {
		t.Fatal("could not read Go version:", err)
	}
	if minor >= 17 {
		tests = append(tests, testCase{"go1.17.go", ""})
	}

	for _, tc := range tests {
		name := tc.file
		targetString := "wasm"
		if tc.target != "" {
			targetString = tc.target
			name = tc.file + "-" + tc.target
		}

		t.Run(name, func(t *testing.T) {
			options := &compileopts.Options{
				Target: targetString,
			}
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
				GOOS:               config.GOOS(),
				GOARCH:             config.GOARCH(),
				CodeModel:          config.CodeModel(),
				RelocationModel:    config.RelocationModel(),
				Scheduler:          config.Scheduler(),
				FuncImplementation: config.FuncImplementation(),
				AutomaticStackSize: config.AutomaticStackSize(),
			}
			machine, err := NewTargetMachine(compilerConfig)
			if err != nil {
				t.Fatal("failed to create target machine:", err)
			}

			// Load entire program AST into memory.
			lprogram, err := loader.Load(config, []string{"./testdata/" + tc.file}, config.ClangHeaders, types.Config{
				Sizes: Sizes(machine),
			})
			if err != nil {
				t.Fatal("failed to create target machine:", err)
			}
			err = lprogram.Parse()
			if err != nil {
				t.Fatalf("could not parse test case %s: %s", tc.file, err)
			}

			// Compile AST to IR.
			program := lprogram.LoadSSA()
			pkg := lprogram.MainPkg()
			mod, errs := CompilePackage(tc.file, pkg, program.Package(pkg.Pkg), machine, compilerConfig, false)
			if errs != nil {
				for _, err := range errs {
					t.Error(err)
				}
				return
			}

			err = llvm.VerifyModule(mod, llvm.PrintMessageAction)
			if err != nil {
				t.Error(err)
			}

			// Optimize IR a little.
			funcPasses := llvm.NewFunctionPassManagerForModule(mod)
			defer funcPasses.Dispose()
			funcPasses.AddInstructionCombiningPass()
			funcPasses.InitializeFunc()
			for fn := mod.FirstFunction(); !fn.IsNil(); fn = llvm.NextFunction(fn) {
				funcPasses.RunFunc(fn)
			}
			funcPasses.FinalizeFunc()

			outFilePrefix := tc.file[:len(tc.file)-3]
			if tc.target != "" {
				outFilePrefix += "-" + tc.target
			}
			outPath := "./testdata/" + outFilePrefix + ".ll"

			// Update test if needed. Do not check the result.
			if *flagUpdate {
				err := ioutil.WriteFile(outPath, []byte(mod.String()), 0666)
				if err != nil {
					t.Error("failed to write updated output file:", err)
				}
				return
			}

			expected, err := ioutil.ReadFile(outPath)
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

package compiler

import (
	"flag"
	"go/types"
	"io/ioutil"
	"strconv"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/loader"
	"tinygo.org/x/go-llvm"
)

// Pass -update to go test to update the output of the test files.
var flagUpdate = flag.Bool("update", false, "update tests based on test output")

// Basic tests for the compiler. Build some Go files and compare the output with
// the expected LLVM IR for regression testing.
func TestCompiler(t *testing.T) {
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

	target, err := compileopts.LoadTarget("i686--linux")
	if err != nil {
		t.Fatal("failed to load target:", err)
	}
	config := &compileopts.Config{
		Options: &compileopts.Options{},
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

	tests := []string{
		"basic.go",
		"pointer.go",
		"slice.go",
		"string.go",
		"float.go",
		"interface.go",
		"func.go",
	}

	for _, testCase := range tests {
		t.Run(testCase, func(t *testing.T) {
			// Load entire program AST into memory.
			lprogram, err := loader.Load(config, []string{"./testdata/" + testCase}, config.ClangHeaders, types.Config{
				Sizes: Sizes(machine),
			})
			if err != nil {
				t.Fatal("failed to create target machine:", err)
			}
			err = lprogram.Parse()
			if err != nil {
				t.Fatalf("could not parse test case %s: %s", testCase, err)
			}

			// Compile AST to IR.
			program := lprogram.LoadSSA()
			pkg := lprogram.MainPkg()
			mod, errs := CompilePackage(testCase, pkg, program.Package(pkg.Pkg), machine, compilerConfig, false)
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

			outfile := "./testdata/" + testCase[:len(testCase)-3] + ".ll"

			// Update test if needed. Do not check the result.
			if *flagUpdate {
				err := ioutil.WriteFile(outfile, []byte(mod.String()), 0666)
				if err != nil {
					t.Error("failed to write updated output file:", err)
				}
				return
			}

			expected, err := ioutil.ReadFile(outfile)
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

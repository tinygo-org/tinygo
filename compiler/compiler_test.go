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
	file      string
	target    string
	scheduler string
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

	// Determine Go minor version (e.g. 16 in go1.16.3).
	_, goMinor, err := goenv.GetGorootVersion(goenv.Get("GOROOT"))
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
		{"pragma.go", "", ""},
		{"goroutine.go", "wasm", "asyncify"},
		{"goroutine.go", "cortex-m-qemu", "tasks"},
		{"channel.go", "", ""},
		{"intrinsics.go", "cortex-m-qemu", ""},
		{"intrinsics.go", "wasm", ""},
		{"gc.go", "", ""},
	}
	if llvmMajor >= 12 {
		tests = append(tests, testCase{"intrinsics.go", "cortex-m-qemu", ""})
		tests = append(tests, testCase{"intrinsics.go", "wasm", ""})
	}
	if goMinor >= 17 {
		tests = append(tests, testCase{"go1.17.go", "", ""})
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
			options := &compileopts.Options{
				Target: targetString,
			}
			target, err := compileopts.LoadTarget(options)
			if err != nil {
				t.Fatal("failed to load target:", err)
			}
			if tc.scheduler != "" {
				options.Scheduler = tc.scheduler
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
				AutomaticStackSize: config.AutomaticStackSize(),
				DefaultStackSize:   config.Target.DefaultStackSize,
				NeedsStackObjects:  config.NeedsStackObjects(),
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
			if tc.scheduler != "" {
				outFilePrefix += "-" + tc.scheduler
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
		if llvmVersion < 12 && strings.HasPrefix(line, "attributes ") {
			// Ignore attribute groups. These may change between LLVM versions.
			// Right now test outputs are for LLVM 12 and higher.
			continue
		}
		if llvmVersion < 13 && strings.HasPrefix(line, "target datalayout = ") {
			// The datalayout string may vary betewen LLVM versions.
			// Right now test outputs are for LLVM 13 and higher.
			continue
		}
		out = append(out, line)
	}
	return out
}

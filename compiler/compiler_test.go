package compiler

import (
	"flag"
	"go/types"
	"io/ioutil"
	"regexp"
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
	target, err := compileopts.LoadTarget("i686--linux")
	if err != nil {
		t.Fatal("failed to load target:", err)
	}
	config := &compileopts.Config{
		Options: &compileopts.Options{},
		Target:  target,
	}
	machine, err := NewTargetMachine(config)
	if err != nil {
		t.Fatal("failed to create target machine:", err)
	}

	tests := []string{
		"basic.go",
		"pointer.go",
		"slice.go",
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
			pkg := lprogram.MainPkg()
			mod, errs := CompilePackage(testCase, pkg, machine, config)
			if errs != nil {
				for _, err := range errs {
					t.Log("error:", err)
				}
				return
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

var alignRegexp = regexp.MustCompile(", align [0-9]+$")

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
		match1 := alignRegexp.MatchString(line1)
		match2 := alignRegexp.MatchString(line2)
		if match1 != match2 {
			// Only one of the lines has the align keyword. Remove it.
			// This is a change to make the test work in both LLVM 10 and LLVM
			// 11 (LLVM 11 appears to automatically add alignment everywhere).
			line1 = alignRegexp.ReplaceAllString(line1, "")
			line2 = alignRegexp.ReplaceAllString(line2, "")
		}
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
		if llvmVersion < 10 && strings.HasPrefix(line, "attributes ") {
			// Ignore attribute groups. These may change between LLVM versions.
			// Right now test outputs are for LLVM 10.
			continue
		}
		if llvmVersion < 10 && strings.HasPrefix(line, "target datalayout ") {
			// Ignore the target layout. This may change between LLVM versions.
			continue
		}
		out = append(out, line)
	}
	return out
}

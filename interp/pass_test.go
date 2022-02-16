package interp

import (
	"flag"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"testing"

	"tinygo.org/x/go-llvm"
)

var update = flag.Bool("update", false, "update interp package tests")

func TestPass(t *testing.T) {
	t.Parallel()

	cases := []string{
		"basic",
	}
	for _, c := range cases {
		c := c
		t.Run(c, func(t *testing.T) {
			t.Parallel()

			// Read the input IR.
			ctx := llvm.NewContext()
			in := filepath.Join("testdata", c+".ll")
			buf, err := llvm.NewMemoryBufferFromFile(in)
			os.Stat(in) // make sure this file is tracked by `go test` caching
			if err != nil {
				t.Fatalf("could not read file: %v", err)
			}
			mod, err := ctx.ParseIR(buf)
			if err != nil {
				t.Fatalf("could not load module:\n%v", err)
			}

			err = Run(mod, mod.NamedFunction("runtime.initAll"))
			if err != nil {
				t.Fatalf("could not run:\n%v", err)
			}

			err = llvm.VerifyModule(mod, llvm.PrintMessageAction)
			if err != nil {
				t.Errorf("IR verification error: %s", err.Error())
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

			out := filepath.Join("testdata", c+".out.ll")
			if *update {
				if err != nil {
					return
				}
				err = ioutil.WriteFile(out, []byte(mod.String()), 0666)
				if err != nil {
					t.Errorf("failed to write updated module: %v", err)
				}
			} else {
				expected, err := ioutil.ReadFile(out)
				if err != nil {
					t.Fatal("failed to read golden file:", err)
				}

				if !fuzzyEqualIR(mod.String(), string(expected)) {
					t.Errorf("output does not match expected output:\n%s", mod.String())
				}
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

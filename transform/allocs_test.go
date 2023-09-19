package transform_test

import (
	"go/token"
	"os"
	"path/filepath"
	"regexp"
	"sort"
	"strconv"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestAllocs(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/allocs", func(mod llvm.Module) {
		transform.OptimizeAllocs(mod, nil, nil)
	})
}

type allocsTestOutput struct {
	filename string
	line     int
	msg      string
}

func (out allocsTestOutput) String() string {
	return out.filename + ":" + strconv.Itoa(out.line) + ": " + out.msg
}

// Test with a Go file as input (for more accurate tests).
func TestAllocs2(t *testing.T) {
	t.Parallel()

	mod := compileGoFileForTesting(t, "./testdata/allocs2.go")

	// Run functionattrs pass, which is necessary for escape analysis.
	po := llvm.NewPassBuilderOptions()
	defer po.Dispose()
	err := mod.RunPasses("function(instcombine),function-attrs", llvm.TargetMachine{}, po)
	if err != nil {
		t.Error("failed to run passes:", err)
	}

	// Run heap to stack transform.
	var testOutputs []allocsTestOutput
	transform.OptimizeAllocs(mod, regexp.MustCompile("."), func(pos token.Position, msg string) {
		testOutputs = append(testOutputs, allocsTestOutput{
			filename: filepath.Base(pos.Filename),
			line:     pos.Line,
			msg:      msg,
		})
	})
	sort.Slice(testOutputs, func(i, j int) bool {
		return testOutputs[i].line < testOutputs[j].line
	})
	testOutput := ""
	for _, out := range testOutputs {
		testOutput += out.String() + "\n"
	}

	// Load expected test output (the OUT: lines).
	testInput, err := os.ReadFile("./testdata/allocs2.go")
	if err != nil {
		t.Fatal("could not read test input:", err)
	}
	var expectedTestOutput string
	for i, line := range strings.Split(strings.ReplaceAll(string(testInput), "\r\n", "\n"), "\n") {
		if idx := strings.Index(line, " // OUT: "); idx > 0 {
			msg := line[idx+len(" // OUT: "):]
			expectedTestOutput += "allocs2.go:" + strconv.Itoa(i+1) + ": " + msg + "\n"
		}
	}

	if testOutput != expectedTestOutput {
		t.Errorf("output does not match expected output:\n%s", testOutput)
	}
}

package transform_test

import (
	"fmt"
	"go/token"
	"os"
	"regexp"
	"sort"
	"strings"
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestAllocs(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/allocs", func(mod llvm.Module) {
		transform.OptimizeAllocs(mod, nil, 256, nil)
	})
}

// Test with a Go file as input (for more accurate tests).
func TestAllocs2(t *testing.T) {
	t.Parallel()

	const (
		basePath   = "testdata/allocs2"
		goFile     = basePath + ".go"
		goldenFile = basePath + ".out"
	)
	mod := compileGoFileForTesting(t, goFile)

	// Run functionattrs pass, which is necessary for escape analysis.
	po := llvm.NewPassBuilderOptions()
	defer po.Dispose()
	err := mod.RunPasses("function(instcombine),function-attrs", llvm.TargetMachine{}, po)
	if err != nil {
		t.Error("failed to run passes:", err)
	}

	// Run heap to stack transform.
	type report struct {
		pos    token.Position
		reason string
	}
	var reports []report
	transform.OptimizeAllocs(mod, regexp.MustCompile("."), 256, func(pos token.Position, reason string) {
		pos.Filename = goFile
		reports = append(reports, report{pos, reason})
	})
	sort.Slice(reports, func(i, j int) bool { return reports[i].pos.Line < reports[j].pos.Line })

	// Load expected test output (the OUT: lines).
	testInput, err := os.ReadFile("./testdata/allocs2.go")
	if err != nil {
		t.Fatal("could not read test input:", err)
	}
	var expectedTestOutput strings.Builder
	for i, line := range strings.Split(strings.ReplaceAll(string(testInput), "\r\n", "\n"), "\n") {
		const prefix = " // OUT: "
		if idx := strings.Index(line, prefix); idx > 0 {
			msg := line[idx+len(prefix):]
			fmt.Fprintf(&expectedTestOutput, "allocs2.go:%d: %s\n", i+1, msg)
		}
	}

	// Check whether the '// OUT' lines in allocs2.go match with the output we
	// got from the test.
	var actualTestOutput strings.Builder
	for _, r := range reports {
		fmt.Fprintf(&actualTestOutput, "allocs2.go:%d: %s\n", r.pos.Line, r.reason)
	}
	if actualTestOutput.String() != expectedTestOutput.String() {
		t.Errorf("expected:\n%s\nactual:\n%s", expectedTestOutput.String(), actualTestOutput.String())
	}

	// Render the cover report and diff it against its golden file.
	var got strings.Builder
	for _, r := range reports {
		if line := transform.FormatAllocCover(r.pos); line != "" {
			got.WriteString(line)
			got.WriteByte('\n')
		}
	}
	checkGolden(t, goldenFile+".cover", got.String())
}

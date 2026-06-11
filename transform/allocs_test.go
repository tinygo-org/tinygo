package transform_test

import (
	"go/token"
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

	// Render every report in each format and diff against its golden file.
	for _, format := range []struct {
		name   string
		render func(report) string
	}{
		{"reason", func(r report) string { return transform.FormatAllocReason(r.pos, r.reason) }},
		{"cover", func(r report) string { return transform.FormatAllocCover(r.pos) }},
	} {
		t.Run(format.name, func(t *testing.T) {
			var got strings.Builder
			for _, r := range reports {
				if line := format.render(r); line != "" {
					got.WriteString(line)
					got.WriteByte('\n')
				}
			}
			checkGolden(t, goldenFile+"."+format.name, got.String())
		})
	}
}

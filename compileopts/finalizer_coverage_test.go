package compileopts

import (
	"go/build/constraint"
	"os"
	"path/filepath"
	"strings"
	"testing"
)

// TestFinalizerRunnerSchedulerCoverage checks that the build constraints on the
// gc_finalizer_sched*.go files define spawnFinalizerRunner for exactly one file
// per scheduler. The three constraints must partition the scheduler space: every
// scheduler matches exactly one file, so none can be left with the symbol
// undefined or defined twice. It iterates validSchedulerOptions as the source of
// truth, so a newly added scheduler is covered by this check automatically.
func TestFinalizerRunnerSchedulerCoverage(t *testing.T) {
	files := []string{
		"gc_finalizer_sched.go",
		"gc_finalizer_sched_none.go",
		"gc_finalizer_sched_other.go",
	}
	exprs := make([]constraint.Expr, len(files))
	for i, name := range files {
		exprs[i] = readBuildConstraint(t, filepath.Join("..", "src", "runtime", name))
	}

	for _, sched := range validSchedulerOptions {
		// The finalizer table exists under the block GCs; gc.conservative
		// satisfies the "gc.conservative || gc.precise" half of every constraint.
		tags := map[string]bool{
			"gc.conservative":    true,
			"scheduler." + sched: true,
		}
		var matched []string
		for i, expr := range exprs {
			if expr.Eval(func(tag string) bool { return tags[tag] }) {
				matched = append(matched, files[i])
			}
		}
		if len(matched) != 1 {
			t.Errorf("scheduler.%s: spawnFinalizerRunner defined in %d files %v, want exactly 1",
				sched, len(matched), matched)
		}
	}
}

// readBuildConstraint returns the parsed //go:build expression of a Go file.
func readBuildConstraint(t *testing.T, path string) constraint.Expr {
	t.Helper()
	data, err := os.ReadFile(path)
	if err != nil {
		t.Fatal(err)
	}
	for _, line := range strings.Split(string(data), "\n") {
		line = strings.TrimSpace(line)
		if constraint.IsGoBuild(line) {
			expr, err := constraint.Parse(line)
			if err != nil {
				t.Fatalf("%s: %v", path, err)
			}
			return expr
		}
		if line != "" && !strings.HasPrefix(line, "//") {
			break // reached code before any //go:build line
		}
	}
	t.Fatalf("%s: no //go:build line found", path)
	return nil
}

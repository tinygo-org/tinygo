package transform_test

import (
	"go/token"
	"regexp"
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestPrintBoundsChecks(t *testing.T) {
	t.Parallel()
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	buf, err := llvm.NewMemoryBufferFromFile("testdata/bounds.ll")
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()

	count := func(re string) int {
		n := 0
		transform.PrintBoundsChecks(mod, regexp.MustCompile(re), func(token.Position, string) { n++ })
		return n
	}
	if got := count(`main\.get`); got != 1 {
		t.Errorf("main.get: got %d bounds checks, want 1", got)
	}
	if got := count(`main\.other`); got != 0 {
		t.Errorf("main.other: got %d bounds checks, want 0", got)
	}
}

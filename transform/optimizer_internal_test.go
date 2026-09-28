package transform

import (
	"os"
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestBlockGlobalAllocPromotionUses(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()
	ensureTestCacheFreshness(t, "testdata/optimizer-alloc-uses.ll")
	buf, err := llvm.NewMemoryBufferFromFile("testdata/optimizer-alloc-uses.ll")
	if err != nil {
		t.Fatal(err)
	}
	mod, err := ctx.ParseIR(buf)
	if err != nil {
		t.Fatal(err)
	}
	defer mod.Dispose()

	blockGlobalAllocPromotion(mod)
	if err := llvm.VerifyModule(mod, llvm.ReturnStatusAction); err != nil {
		t.Fatal(err)
	}
	marker := mod.NamedFunction("tinygo.gc.alloc.marker")
	if marker.IsNil() {
		t.Fatal("allocation marker was not created")
	}
	if uses := getUses(marker); len(uses) != 1 {
		t.Fatalf("got %d marker uses, want 1", len(uses))
	}
}

// ensureTestCacheFreshness registers path as an input of the running test.
// see https://github.com/tinygo-org/tinygo/issues/5780.
func ensureTestCacheFreshness(t *testing.T, path string) {
	t.Helper()
	if _, err := os.Stat(path); err != nil {
		t.Fatalf("could not stat test fixture %s: %v", path, err)
	}
}

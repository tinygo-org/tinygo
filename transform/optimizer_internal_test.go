package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestBlockGlobalAllocPromotionUses(t *testing.T) {
	ctx := llvm.NewContext()
	defer ctx.Dispose()
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

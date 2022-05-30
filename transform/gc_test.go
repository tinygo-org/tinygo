package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestMakeGCStackSlots(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/gc-stackslots", func(mod llvm.Module) {
		transform.MakeGCStackSlots(mod)
	})
}

package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestAddGlobalsBitmap(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/gc-globals", func(mod llvm.Module) {
		transform.AddGlobalsBitmap(mod)
	})
}

func TestMakeGCStackSlots(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/gc-stackslots", func(mod llvm.Module) {
		transform.MakeGCStackSlots(mod)
	})
}

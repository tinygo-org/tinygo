package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestAddGlobalsBitmap(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/gc-globals", func(mod llvm.Module) {
		AddGlobalsBitmap(mod)
	})
}

func TestMakeGCStackSlots(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/gc-stackslots", func(mod llvm.Module) {
		MakeGCStackSlots(mod)
	})
}

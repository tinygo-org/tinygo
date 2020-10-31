package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestApplyFunctionSections(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/globals-function-sections", func(mod llvm.Module) {
		ApplyFunctionSections(mod)
	})
}

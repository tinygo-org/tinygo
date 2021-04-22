package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestApplyFunctionSections(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/globals-function-sections", func(mod llvm.Module) {
		transform.ApplyFunctionSections(mod)
	})
}

package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestOptimizeStringToBytes(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringtobytes", func(mod llvm.Module) {
		// Run optimization pass.
		transform.OptimizeStringToBytes(mod)
	})
}

func TestOptimizeStringEqual(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringequal", func(mod llvm.Module) {
		// Run optimization pass.
		transform.OptimizeStringEqual(mod)
	})
}

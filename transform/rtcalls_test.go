package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestOptimizeStringToBytes(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringtobytes", func(mod llvm.Module) {
		// Run optimization pass.
		OptimizeStringToBytes(mod)
	})
}

func TestOptimizeStringEqual(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringequal", func(mod llvm.Module) {
		// Run optimization pass.
		OptimizeStringEqual(mod)
	})
}

func TestOptimizeReflectImplements(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/reflect-implements", func(mod llvm.Module) {
		// Run optimization pass.
		OptimizeReflectImplements(mod)
	})
}

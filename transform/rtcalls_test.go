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

func TestOptimizeStringFromBytesStringEqual(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringfrombytes-stringequal", func(mod llvm.Module) {
		transform.OptimizeStringFromBytes(mod)
	})
}

func TestOptimizeStringFromBytesStringLess(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringfrombytes-stringless", func(mod llvm.Module) {
		transform.OptimizeStringFromBytes(mod)
	})
}

func TestOptimizeStringFromBytesLen(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stringfrombytes-len", func(mod llvm.Module) {
		transform.OptimizeStringFromBytes(mod)
	})
}

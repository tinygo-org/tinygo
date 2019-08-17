package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestOptimizeMaps(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/maps", func(mod llvm.Module) {
		// Run optimization pass.
		OptimizeMaps(mod)

		// Run an optimization pass, to clean up the result.
		// This shows that all code related to the map is really eliminated.
		pm := llvm.NewPassManager()
		defer pm.Dispose()
		pm.AddDeadStoreEliminationPass()
		pm.Run(mod)
	})
}

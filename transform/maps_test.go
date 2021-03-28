package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestOptimizeMaps(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/maps", func(mod llvm.Module) {
		// Run optimization pass.
		transform.OptimizeMaps(mod)

		// Run an optimization pass, to clean up the result.
		// This shows that all code related to the map is really eliminated.
		pm := llvm.NewPassManager()
		defer pm.Dispose()
		pm.AddDeadStoreEliminationPass()
		pm.AddAggressiveDCEPass()
		pm.Run(mod)
	})
}

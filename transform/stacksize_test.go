package transform

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"tinygo.org/x/go-llvm"
)

func TestCreateStackSizeLoads(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stacksize", func(mod llvm.Module) {
		// Run optimization pass.
		CreateStackSizeLoads(mod, &compileopts.Config{
			Target: &compileopts.TargetSpec{
				DefaultStackSize: 1024,
			},
		})
	})
}

package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestCreateStackSizeLoads(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/stacksize", func(mod llvm.Module) {
		// Run optimization pass.
		_, err := transform.CreateStackSizeLoads(mod, &compileopts.Config{
			Options: &compileopts.Options{},
			Target: &compileopts.TargetSpec{
				DefaultStackSize: 1024,
			},
		})
		if err != nil {
			t.Error(err)
		}
	})
}

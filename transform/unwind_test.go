package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestUnwindAssumptions(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/unwind", func(mod llvm.Module) {
		transform.AddUnwindAssumptions(mod)
		po := llvm.NewPassBuilderOptions()
		defer po.Dispose()
		if err := mod.RunPasses("thinlto-pre-link<Oz>", llvm.TargetMachine{}, po); err != nil {
			t.Fatal(err)
		}
	})
}

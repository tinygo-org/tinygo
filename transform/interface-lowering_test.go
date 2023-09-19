package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestInterfaceLowering(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/interface", func(mod llvm.Module) {
		err := transform.LowerInterfaces(mod, defaultTestConfig)
		if err != nil {
			t.Error(err)
		}

		po := llvm.NewPassBuilderOptions()
		defer po.Dispose()
		err = mod.RunPasses("globaldce", llvm.TargetMachine{}, po)
		if err != nil {
			t.Error("failed to run passes:", err)
		}
	})
}

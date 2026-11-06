package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/compiler/llvmutil"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestUnwindAssumptions(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/unwind", func(mod llvm.Module) {
		transform.AddUnwindAssumptions(mod)
		po := llvm.NewPassBuilderOptions()
		defer po.Dispose()
		// LLVM 23 removed the Oz pipeline level. See transform/optimizer.go.
		level := "Oz"
		if llvmutil.Version() >= 23 {
			level = "O2"
		}
		if err := mod.RunPasses("thinlto-pre-link<"+level+">", llvm.TargetMachine{}, po); err != nil {
			t.Fatal(err)
		}
	})
}

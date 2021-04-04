package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestInterfaceLowering(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/interface", func(mod llvm.Module) {
		err := LowerInterfaces(mod, 0)
		if err != nil {
			t.Error(err)
		}
	})
}

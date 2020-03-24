package transform

import (
	"testing"
	"tinygo.org/x/go-llvm"
)

func TestGoroutineLowering(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/coroutines", func(mod llvm.Module) {
		err := LowerCoroutines(mod, false)
		if err != nil {
			panic(err)
		}
	})
}

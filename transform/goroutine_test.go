package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestGoroutineLowering(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/coroutines", func(mod llvm.Module) {
		err := transform.LowerCoroutines(mod, false)
		if err != nil {
			panic(err)
		}
	})
}

package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestInterruptLowering(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/interrupt", func(mod llvm.Module) {
		errs := transform.LowerInterrupts(mod)
		if len(errs) != 0 {
			t.Fail()
			for _, err := range errs {
				t.Error(err)
			}
		}
	})
}

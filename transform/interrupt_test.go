package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestInterruptLowering(t *testing.T) {
	t.Parallel()
	for _, subtest := range []string{"avr", "cortexm"} {
		t.Run(subtest, func(t *testing.T) {
			testTransform(t, "testdata/interrupt-"+subtest, func(mod llvm.Module) {
				errs := LowerInterrupts(mod, 0)
				if len(errs) != 0 {
					t.Fail()
					for _, err := range errs {
						t.Error(err)
					}
				}
			})
		})
	}
}

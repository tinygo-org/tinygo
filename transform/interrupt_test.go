package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestInterruptLowering(t *testing.T) {
	t.Parallel()
	for _, subtest := range []string{"avr", "cortexm"} {
		t.Run(subtest, func(t *testing.T) {
			testTransform(t, "testdata/interrupt-"+subtest, func(mod llvm.Module) {
				errs := transform.LowerInterrupts(mod, &compileopts.Config{Options: &compileopts.Options{Opt: "2"}})
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

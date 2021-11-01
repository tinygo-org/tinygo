package transform_test

import (
	"testing"

	"github.com/tinygo-org/tinygo/compileopts"
	"github.com/tinygo-org/tinygo/transform"
	"tinygo.org/x/go-llvm"
)

func TestWasmABI(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/wasm-abi", func(mod llvm.Module) {
		// Run ABI change pass.
		err := transform.ExternalInt64AsPtr(mod, &compileopts.Config{Options: &compileopts.Options{Opt: "2"}})
		if err != nil {
			t.Errorf("failed to change wasm ABI: %v", err)
		}
	})
}

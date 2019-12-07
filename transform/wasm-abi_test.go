package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestWasmABI(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/wasm-abi", func(mod llvm.Module) {
		// Run ABI change pass.
		err := ExternalInt64AsPtr(mod)
		if err != nil {
			t.Errorf("failed to change wasm ABI: %v", err)
		}
	})
}

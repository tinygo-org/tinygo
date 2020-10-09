package transform

import (
	"testing"

	"tinygo.org/x/go-llvm"
)

func TestAddWasmCustomSections(t *testing.T) {
	t.Parallel()
	testTransform(t, "testdata/wasm-custom-sections", func(mod llvm.Module) {
		AddWASMCustomSections(mod, [][2]string{
			{"red", "foo"},
			{"green", "bar"},
			{"green", "qux"},
		})
	})
}

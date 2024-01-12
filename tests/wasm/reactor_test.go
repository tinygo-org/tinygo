package wasm

import (
	"strings"
	"testing"
)

func TestReactor(t *testing.T) {
	tmpDir := t.TempDir()

	err := run(t, "tinygo build -x -o "+tmpDir+"/reactor.wasm -target wasi testdata/reactor.go")
	if err != nil {
		t.Fatal(err)
	}

	out, err := runout(t, "wasmtime run --invoke tinygo_test "+tmpDir+"/reactor.wasm")
	if err != nil {
		t.Fatal(err)
	}

	got := string(out)
	want := "1337\n"
	if !strings.Contains(got, want) {
		t.Errorf("reactor: expected %s, got %s", want, got)
	}
}

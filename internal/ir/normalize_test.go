package ir

import "testing"

func TestNormalizeWasiTriple(t *testing.T) {
	for _, tc := range []struct {
		in, want string
	}{
		{`target triple = "wasm32-unknown-wasi"`, `target triple = "wasm32-unknown-wasip1"`},
		{`target triple = "wasm32-unknown-wasip1"`, `target triple = "wasm32-unknown-wasip1"`},
		{`target triple = "wasm32-unknown-wasip2"`, `target triple = "wasm32-unknown-wasip2"`},
		{`target triple = "wasm32-unknown-unknown"`, `target triple = "wasm32-unknown-unknown"`},
	} {
		if got := Normalize(tc.in); got != tc.want {
			t.Errorf("Normalize(%q) = %q, want %q", tc.in, got, tc.want)
		}
	}
}

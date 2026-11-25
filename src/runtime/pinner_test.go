package runtime_test

import (
	"runtime"
	"testing"
	"unsafe"
)

func TestPinner(t *testing.T) {
	var p runtime.Pinner
	p.Unpin()

	x := new(int)
	p.Pin(x)
	p.Pin(unsafe.Pointer(x))
	p.Pin((*int)(nil))
	p.Pin(unsafe.Pointer(nil))
	p.Unpin()

	// Reuse after Unpin.
	p.Pin(x)
	p.Unpin()
}

func TestPinnerPanics(t *testing.T) {
	for _, tc := range []struct {
		name string
		arg  any
	}{
		{"nil", nil},
		{"int", 42},
		{"string", "abc"},
	} {
		t.Run(tc.name, func(t *testing.T) {
			defer func() {
				if recover() == nil {
					t.Fatalf("Pin(%v) did not panic", tc.arg)
				}
			}()
			var p runtime.Pinner
			p.Pin(tc.arg)
		})
	}
}

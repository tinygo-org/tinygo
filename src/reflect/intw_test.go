//go:build !avr

package reflect_test

import (
	"reflect"
	"testing"
)

// TestSliceHeaderIntegerSize verifies that SliceHeader.Len and Cap are type int on non-AVR platforms.
// See https://github.com/tinygo-org/tinygo/issues/1284.
func TestSliceHeaderIntegerSize(t *testing.T) {
	var h reflect.SliceHeader
	h.Len = int(0)
	h.Cap = int(0)
}

// TestStringHeaderIntegerSize verifies that StringHeader.Len and Cap are type int on non-AVR platforms.
// See https://github.com/tinygo-org/tinygo/issues/1284.
func TestStringHeaderIntegerSize(t *testing.T) {
	var h reflect.StringHeader
	h.Len = int(0)
}

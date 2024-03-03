//go:build avr

package reflect_test

import (
	"reflect"
	"testing"
	"unsafe"
)

// Verify that SliceHeader is the same size as a slice.
var _ [unsafe.Sizeof([]byte{})]byte = [unsafe.Sizeof(reflect.SliceHeader{})]byte{}

// TestSliceHeaderIntegerSize verifies that SliceHeader.Len and Cap are type uintptr on AVR platforms.
// See https://github.com/tinygo-org/tinygo/issues/1284.
func TestSliceHeaderIntegerSize(t *testing.T) {
	var h reflect.SliceHeader
	h.Len = uintptr(0)
	h.Cap = uintptr(0)
}

// Verify that StringHeader is the same size as a string.
var _ [unsafe.Sizeof("hello")]byte = [unsafe.Sizeof(reflect.StringHeader{})]byte{}

// TestStringHeaderIntegerSize verifies that StringHeader.Len and Cap are type uintptr on AVR platforms.
// See https://github.com/tinygo-org/tinygo/issues/1284.
func TestStringHeaderIntegerSize(t *testing.T) {
	var h reflect.StringHeader
	h.Len = uintptr(0)
}

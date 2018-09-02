package runtime

// This file implements functions related to Go strings.

import (
	"unsafe"
)

// The underlying struct for the Go string type.
type _string struct {
	length lenType
	ptr    *byte
}

// Return true iff the strings match.
//go:nobounds
func stringEqual(x, y string) bool {
	if len(x) != len(y) {
		return false
	}
	for i := 0; i < len(x); i++ {
		if x[i] != y[i] {
			return false
		}
	}
	return true
}

// Add two strings together.
func stringConcat(x, y _string) _string {
	if x.length == 0 {
		return y
	} else if y.length == 0 {
		return x
	} else {
		length := uintptr(x.length + y.length)
		buf := alloc(length)
		memcpy(buf, unsafe.Pointer(x.ptr), uintptr(x.length))
		memcpy(unsafe.Pointer(uintptr(buf)+uintptr(x.length)), unsafe.Pointer(y.ptr), uintptr(y.length))
		return _string{lenType(length), (*byte)(buf)}
	}
}

// Create a string from a []byte slice.
func stringFromBytes(x []byte) _string {
	buf := alloc(uintptr(len(x)))
	for i, c := range x {
		*(*byte)(unsafe.Pointer(uintptr(buf) + uintptr(i))) = c
	}
	return _string{lenType(len(x)), (*byte)(buf)}
}

// Create a string from a Unicode code point.
func stringFromUnicode(x rune) _string {
	array, length := encodeUTF8(x)
	// Array will be heap allocated.
	// The heap most likely doesn't work with blocks below 4 bytes, so there's
	// no point in allocating a smaller buffer for the string here.
	return _string{length, (*byte)(unsafe.Pointer(&array))}
}

// Convert a Unicode code point into an array of bytes and its length.
func encodeUTF8(x rune) ([4]byte, lenType) {
	// https://stackoverflow.com/questions/6240055/manually-converting-unicode-codepoints-into-utf-8-and-utf-16
	// Note: this code can probably be optimized (in size and speed).
	switch {
	case x <= 0x7f:
		return [4]byte{byte(x), 0, 0, 0}, 1
	case x <= 0x7ff:
		b1 := 0xc0 | byte(x>>6)
		b2 := 0x80 | byte(x&0x3f)
		return [4]byte{b1, b2, 0, 0}, 2
	case x <= 0xffff:
		b1 := 0xe0 | byte(x>>12)
		b2 := 0x80 | byte((x>>6)&0x3f)
		b3 := 0x80 | byte((x>>0)&0x3f)
		return [4]byte{b1, b2, b3, 0}, 3
	case x <= 0x10ffff:
		b1 := 0xf0 | byte(x>>18)
		b2 := 0x80 | byte((x>>12)&0x3f)
		b3 := 0x80 | byte((x>>6)&0x3f)
		b4 := 0x80 | byte((x>>0)&0x3f)
		return [4]byte{b1, b2, b3, b4}, 4
	default:
		// Invalid Unicode code point.
		return [4]byte{0xef, 0xbf, 0xbd, 0}, 3
	}
}

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

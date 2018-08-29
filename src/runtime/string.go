package runtime

// This file implements functions related to Go strings.

import (
	"unsafe"
)

// The underlying struct for the Go string type.
type _string struct {
	length lenType
	ptr    *uint8
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
		return _string{lenType(length), (*uint8)(buf)}
	}
}

package runtime

// This file implements various core algorithms used in the runtime package and
// standard library.

import (
	"unsafe"
)

// This function is used by hash/maphash.
func fastrand() uint32 {
	xorshift32State = xorshift32(xorshift32State)
	return xorshift32State
}

var xorshift32State uint32 = 1

func xorshift32(x uint32) uint32 {
	// Algorithm "xor" from p. 4 of Marsaglia, "Xorshift RNGs".
	// Improved sequence based on
	// http://www.iro.umontreal.ca/~lecuyer/myftp/papers/xorshift.pdf
	x ^= x << 7
	x ^= x >> 1
	x ^= x << 9
	return x
}

// This function is used by hash/maphash.
func memhash(p unsafe.Pointer, seed, s uintptr) uintptr {
	if unsafe.Sizeof(uintptr(0)) > 4 {
		return uintptr(hash64(p, s, seed))
	}
	return uintptr(hash32(p, s, seed))
}

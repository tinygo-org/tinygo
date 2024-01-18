package runtime

// This file implements various core algorithms used in the runtime package and
// standard library.

import (
	"unsafe"
)

// This function is needed by math/rand since Go 1.20.
// See: https://github.com/golang/go/issues/54880
//
//go:linkname rand_fastrand64 math/rand.fastrand64
func rand_fastrand64() uint64 {
	return fastrand64()
}

// This function is used by hash/maphash.
// This function isn't required anymore since Go 1.22, so should be removed once
// that becomes the minimum requirement.
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
// This function isn't required anymore since Go 1.22, so should be removed once
// that becomes the minimum requirement.
func fastrand64() uint64 {
	xorshift64State = xorshiftMult64(xorshift64State)
	return xorshift64State
}

var xorshift64State uint64 = 1

// 64-bit xorshift multiply rng from http://vigna.di.unimi.it/ftp/papers/xorshift.pdf
func xorshiftMult64(x uint64) uint64 {
	x ^= x >> 12 // a
	x ^= x << 25 // b
	x ^= x >> 27 // c
	return x * 2685821657736338717
}

// This function is used by hash/maphash.
func memhash(p unsafe.Pointer, seed, s uintptr) uintptr {
	if unsafe.Sizeof(uintptr(0)) > 4 {
		return uintptr(hash64(p, s, seed))
	}
	return uintptr(hash32(p, s, seed))
}

// Function that's called from various packages starting with Go 1.22.
func rand() uint64 {
	// Return a random number from hardware, falling back to software if
	// unavailable.
	n, ok := hardwareRand()
	if !ok {
		// Fallback to static random number.
		// Not great, but we can't do much better than this.
		n = fastrand64()
	}
	return n
}

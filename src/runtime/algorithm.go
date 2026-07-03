package runtime

// This file implements various core algorithms used in the runtime package and
// standard library.

import (
	"unsafe"
)

func initRand() {
	r, _ := hardwareRand()
	xorshift64State = uint64(r | 1) // protect against 0
}

var xorshift64State uint64 = 1

// 32-bit xorshift mixer used by the leveldb hash implementation.
func xorshift32(x uint32) uint32 {
	x ^= x << 13
	x ^= x >> 7
	x ^= x << 17
	return x
}

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
	if ok {
		return n
	}

	// Fallback to a deterministic software generator.
	xorshift64State = xorshiftMult64(xorshift64State)
	return xorshift64State
}

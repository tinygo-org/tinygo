package runtime

// This file implements various core algorithms used in the runtime package and
// standard library.

import "unsafe"

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

// Get FNV-1a hash of the given memory buffer.
//
// https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function#FNV-1a_hash
func hash32(ptr unsafe.Pointer, n uintptr, seed uintptr) uint32 {
	var result uint32 = 2166136261 ^ uint32(seed) // FNV offset basis
	for i := uintptr(0); i < n; i++ {
		c := *(*uint8)(unsafe.Pointer(uintptr(ptr) + i))
		result ^= uint32(c) // XOR with byte
		result *= 16777619  // FNV prime
	}
	return result
}

// Also a FNV-1a hash.
func hash64(ptr unsafe.Pointer, n uintptr, seed uintptr) uint64 {
	var result uint64 = 14695981039346656037 ^ uint64(seed) // FNV offset basis
	for i := uintptr(0); i < n; i++ {
		c := *(*uint8)(unsafe.Pointer(uintptr(ptr) + i))
		result ^= uint64(c)     // XOR with byte
		result *= 1099511628211 // FNV prime
	}
	return result
}

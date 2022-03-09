//go:build !runtime_memhash_tsip
// +build !runtime_memhash_tsip

package runtime

import "unsafe"

// Get FNV-1a hash of the given memory buffer.
//
// https://en.wikipedia.org/wiki/Fowler%E2%80%93Noll%E2%80%93Vo_hash_function#FNV-1a_hash
func hash32(ptr unsafe.Pointer, n, seed uintptr) uint32 {
	var result uint32 = 2166136261 // FNV offset basis
	result *= uint32(seed)
	for i := uintptr(0); i < n; i++ {
		c := *(*uint8)(unsafe.Pointer(uintptr(ptr) + i))
		result ^= uint32(c) // XOR with byte
		result *= 16777619  // FNV prime
	}
	return result
}

// Also a FNV-1a hash.
func hash64(ptr unsafe.Pointer, n, seed uintptr) uint64 {
	var result uint64 = 14695981039346656037 // FNV offset basis
	result *= uint64(seed)
	for i := uintptr(0); i < n; i++ {
		c := *(*uint8)(unsafe.Pointer(uintptr(ptr) + i))
		result ^= uint64(c)     // XOR with byte
		result *= 1099511628211 // FNV prime
	}
	return result
}

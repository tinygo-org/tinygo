//go:build runtime_memhash_leveldb
// +build runtime_memhash_leveldb

// This is the hash function from Google's leveldb key-value storage system. It
// processes 4 bytes at a time making it faster than the FNV hash for buffer
// lengths > 16 bytes.

// https://github.com/google/leveldb
// https://en.wikipedia.org/wiki/LevelDB

package runtime

import (
	"unsafe"
)

func ptrToSlice(ptr unsafe.Pointer, n uintptr) []byte {
	var p []byte

	type _bslice struct {
		ptr *byte
		len uintptr
		cap uintptr
	}

	pslice := (*_bslice)(unsafe.Pointer(&p))
	pslice.ptr = (*byte)(ptr)
	pslice.cap = n
	pslice.len = n

	return p
}

// leveldb hash
func hash32(ptr unsafe.Pointer, n, seed uintptr) uint32 {

	const (
		lseed = 0xbc9f1d34
		m     = 0xc6a4a793
	)

	b := ptrToSlice(ptr, n)

	h := uint32(lseed^seed) ^ uint32(uint(len(b))*uint(m))

	for ; len(b) >= 4; b = b[4:] {
		h += uint32(b[0]) | uint32(b[1])<<8 | uint32(b[2])<<16 | uint32(b[3])<<24
		h *= m
		h ^= h >> 16
	}
	switch len(b) {
	case 3:
		h += uint32(b[2]) << 16
		fallthrough
	case 2:
		h += uint32(b[1]) << 8
		fallthrough
	case 1:
		h += uint32(b[0])
		h *= m
		h ^= h >> 24
	}

	return h
}

// hash64finalizer is a 64-bit integer mixing function from
// https://web.archive.org/web/20120720045250/http://www.cris.com/~Ttwang/tech/inthash.htm
func hash64finalizer(key uint64) uint64 {
	key = ^key + (key << 21) // key = (key << 21) - key - 1;
	key = key ^ (key >> 24)
	key = (key + (key << 3)) + (key << 8) // key * 265
	key = key ^ (key >> 14)
	key = (key + (key << 2)) + (key << 4) // key * 21
	key = key ^ (key >> 28)
	key = key + (key << 31)
	return key
}

// hash64 turns hash32 into a 64-bit hash, by use hash64finalizer to
// mix the result of hash32 function combined with an xorshifted version of
// the seed.
func hash64(ptr unsafe.Pointer, n, seed uintptr) uint64 {
	h32 := hash32(ptr, n, seed)
	return hash64finalizer((uint64(h32^xorshift32(uint32(seed))) << 32) | uint64(h32))
}

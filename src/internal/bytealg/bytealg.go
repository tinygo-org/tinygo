package bytealg

const (
	// Index can search any valid length of string.

	MaxLen        = int(-1) >> 31
	MaxBruteForce = MaxLen
)

// Compare two byte slices.
// Returns -1 if the first differing byte is lower in a, or 1 if the first differing byte is greater in b.
// If the byte slices are equal, returns 0.
// If the lengths are different and there are no differing bytes, compares based on length.
func Compare(a, b []byte) int {
	// Compare for differing bytes.
	for i := 0; i < len(a) && i < len(b); i++ {
		switch {
		case a[i] < b[i]:
			return -1
		case a[i] > b[i]:
			return 1
		}
	}

	// Compare lengths.
	switch {
	case len(a) > len(b):
		return 1
	case len(a) < len(b):
		return -1
	default:
		return 0
	}
}

// Count the number of instances of a byte in a slice.
func Count(b []byte, c byte) int {
	// Use a simple implementation, as there is no intrinsic that does this like we want.
	n := 0
	for _, v := range b {
		if v == c {
			n++
		}
	}
	return n
}

// Count the number of instances of a byte in a string.
func CountString(s string, c byte) int {
	// Use a simple implementation, as there is no intrinsic that does this like we want.
	// Currently, the compiler does not generate zero-copy byte-string conversions, so this needs to be seperate from Count.
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}
	return n
}

// Cutover is not reachable in TinyGo, but must exist as it is referenced.
func Cutover(n int) int {
	// Setting MaxLen and MaxBruteForce should force a different path to be taken.
	// This should never be called.
	panic("cutover is unreachable")
}

// Equal checks if two byte slices are equal.
// It is equivalent to bytes.Equal.
func Equal(a, b []byte) bool {
	if len(a) != len(b) {
		return false
	}

	for i, v := range a {
		if v != b[i] {
			return false
		}
	}

	return true
}

// Index finds the base index of the first instance of the byte sequence b in a.
// If a does not contain b, this returns -1.
func Index(a, b []byte) int {
	for i := 0; i <= len(a)-len(b); i++ {
		if Equal(a[i:i+len(b)], b) {
			return i
		}
	}
	return -1
}

// Index finds the index of the first instance of the specified byte in the slice.
// If the byte is not found, this returns -1.
func IndexByte(b []byte, c byte) int {
	for i, v := range b {
		if v == c {
			return i
		}
	}
	return -1
}

// Index finds the index of the first instance of the specified byte in the string.
// If the byte is not found, this returns -1.
func IndexByteString(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// Index finds the base index of the first instance of a substring in a string.
// If the substring is not found, this returns -1.
func IndexString(str, sub string) int {
	for i := 0; i <= len(str)-len(sub); i++ {
		if str[i:i+len(sub)] == sub {
			return i
		}
	}
	return -1
}

// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// The following code has been copied from the Go 1.15 release tree.

// PrimeRK is the prime base used in Rabin-Karp algorithm.
const PrimeRK = 16777619

// HashStrBytes returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
func HashStrBytes(sep []byte) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*PrimeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, PrimeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

// HashStr returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
func HashStr(sep string) (uint32, uint32) {
	hash := uint32(0)
	for i := 0; i < len(sep); i++ {
		hash = hash*PrimeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, PrimeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

// HashStrRevBytes returns the hash of the reverse of sep and the
// appropriate multiplicative factor for use in Rabin-Karp algorithm.
func HashStrRevBytes(sep []byte) (uint32, uint32) {
	hash := uint32(0)
	for i := len(sep) - 1; i >= 0; i-- {
		hash = hash*PrimeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, PrimeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

// HashStrRev returns the hash of the reverse of sep and the
// appropriate multiplicative factor for use in Rabin-Karp algorithm.
func HashStrRev(sep string) (uint32, uint32) {
	hash := uint32(0)
	for i := len(sep) - 1; i >= 0; i-- {
		hash = hash*PrimeRK + uint32(sep[i])
	}
	var pow, sq uint32 = 1, PrimeRK
	for i := len(sep); i > 0; i >>= 1 {
		if i&1 != 0 {
			pow *= sq
		}
		sq *= sq
	}
	return hash, pow
}

// IndexRabinKarpBytes uses the Rabin-Karp search algorithm to return the index of the
// first occurence of substr in s, or -1 if not present.
func IndexRabinKarpBytes(s, sep []byte) int {
	// Rabin-Karp search
	hashsep, pow := HashStrBytes(sep)
	n := len(sep)
	var h uint32
	for i := 0; i < n; i++ {
		h = h*PrimeRK + uint32(s[i])
	}
	if h == hashsep && Equal(s[:n], sep) {
		return 0
	}
	for i := n; i < len(s); {
		h *= PrimeRK
		h += uint32(s[i])
		h -= pow * uint32(s[i-n])
		i++
		if h == hashsep && Equal(s[i-n:i], sep) {
			return i - n
		}
	}
	return -1
}

// IndexRabinKarp uses the Rabin-Karp search algorithm to return the index of the
// first occurence of substr in s, or -1 if not present.
func IndexRabinKarp(s, substr string) int {
	// Rabin-Karp search
	hashss, pow := HashStr(substr)
	n := len(substr)
	var h uint32
	for i := 0; i < n; i++ {
		h = h*PrimeRK + uint32(s[i])
	}
	if h == hashss && s[:n] == substr {
		return 0
	}
	for i := n; i < len(s); {
		h *= PrimeRK
		h += uint32(s[i])
		h -= pow * uint32(s[i-n])
		i++
		if h == hashss && s[i-n:i] == substr {
			return i - n
		}
	}
	return -1
}

// MakeNoZero makes a slice of length and capacity n without zeroing the bytes.
// It is the caller's responsibility to ensure uninitialized bytes
// do not leak to the end user.
func MakeNoZero(n int) []byte {
	// Note: this does zero the buffer even though that's not necessary.
	// For performance reasons we might want to change this (similar to the
	// malloc function implemented in the runtime).
	return make([]byte, n)
}

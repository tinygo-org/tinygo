package bytealg

// Some code in this file has been copied from the Go source code, and has
// copyright of their original authors:
//
// Copyright 2020 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.
//
// This is indicated specifically in the file.

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

// This function was copied from the Go 1.23 source tree (with runtime_cmpstring
// manually inlined).
func CompareString(a, b string) int {
	l := len(a)
	if len(b) < l {
		l = len(b)
	}
	for i := 0; i < l; i++ {
		c1, c2 := a[i], b[i]
		if c1 < c2 {
			return -1
		}
		if c1 > c2 {
			return +1
		}
	}
	if len(a) < len(b) {
		return -1
	}
	if len(a) > len(b) {
		return +1
	}
	return 0
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
	// Currently, the compiler does not generate zero-copy byte-string conversions, so this needs to be separate from Count.
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

// The following code has been copied from the Go 1.15 release tree.

// PrimeRK is the prime base used in Rabin-Karp algorithm.
const PrimeRK = 16777619

// HashStrBytes returns the hash and the appropriate multiplicative
// factor for use in Rabin-Karp algorithm.
//
// This function was removed in Go 1.22.
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
//
// This function was removed in Go 1.22.
func HashStr[T string | []byte](sep T) (uint32, uint32) {
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
//
// This function was removed in Go 1.22.
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
//
// Copied from the Go 1.22rc1 source tree.
func HashStrRev[T string | []byte](sep T) (uint32, uint32) {
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
// first occurrence of substr in s, or -1 if not present.
//
// This function was removed in Go 1.22.
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
// first occurrence of sep in s, or -1 if not present.
//
// Copied from the Go 1.22rc1 source tree.
func IndexRabinKarp[T string | []byte](s, sep T) int {
	// Rabin-Karp search
	hashss, pow := HashStr(sep)
	n := len(sep)
	var h uint32
	for i := 0; i < n; i++ {
		h = h*PrimeRK + uint32(s[i])
	}
	if h == hashss && string(s[:n]) == string(sep) {
		return 0
	}
	for i := n; i < len(s); {
		h *= PrimeRK
		h += uint32(s[i])
		h -= pow * uint32(s[i-n])
		i++
		if h == hashss && string(s[i-n:i]) == string(sep) {
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

// Copied from the Go 1.22rc1 source tree.
func LastIndexByte(s []byte, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// Copied from the Go 1.22rc1 source tree.
func LastIndexByteString(s string, c byte) int {
	for i := len(s) - 1; i >= 0; i-- {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// LastIndexRabinKarp uses the Rabin-Karp search algorithm to return the last index of the
// occurrence of sep in s, or -1 if not present.
//
// Copied from the Go 1.22rc1 source tree.
func LastIndexRabinKarp[T string | []byte](s, sep T) int {
	// Rabin-Karp search from the end of the string
	hashss, pow := HashStrRev(sep)
	n := len(sep)
	last := len(s) - n
	var h uint32
	for i := len(s) - 1; i >= last; i-- {
		h = h*PrimeRK + uint32(s[i])
	}
	if h == hashss && string(s[last:]) == string(sep) {
		return last
	}
	for i := last - 1; i >= 0; i-- {
		h *= PrimeRK
		h += uint32(s[i])
		h -= pow * uint32(s[i+n])
		if h == hashss && string(s[i:i+n]) == string(sep) {
			return i
		}
	}
	return -1
}

package runtime

// This file implements functions related to Go strings.

import (
	"unsafe"
)

// The underlying struct for the Go string type.
type _string struct {
	ptr    *byte
	length uintptr
}

// The iterator state for a range over a string.
type stringIterator struct {
	byteindex uintptr
}

// Return true iff the strings match.
//go:nobounds
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

// Return true iff x < y.
//go:nobounds
func stringLess(x, y string) bool {
	l := len(x)
	if m := len(y); m < l {
		l = m
	}
	for i := 0; i < l; i++ {
		if x[i] < y[i] {
			return true
		}
		if x[i] > y[i] {
			return false
		}
	}
	return len(x) < len(y)
}

// Add two strings together.
func stringConcat(x, y _string) _string {
	if x.length == 0 {
		return y
	} else if y.length == 0 {
		return x
	} else {
		length := x.length + y.length
		buf := alloc(length)
		memcpy(buf, unsafe.Pointer(x.ptr), x.length)
		memcpy(unsafe.Pointer(uintptr(buf)+x.length), unsafe.Pointer(y.ptr), y.length)
		return _string{ptr: (*byte)(buf), length: length}
	}
}

// Create a string from a []byte slice.
func stringFromBytes(x struct {
	ptr *byte
	len uintptr
	cap uintptr
}) _string {
	buf := alloc(x.len)
	memcpy(buf, unsafe.Pointer(x.ptr), x.len)
	return _string{ptr: (*byte)(buf), length: x.len}
}

// Convert a string to a []byte slice.
func stringToBytes(x _string) (slice struct {
	ptr *byte
	len uintptr
	cap uintptr
}) {
	buf := alloc(x.length)
	memcpy(buf, unsafe.Pointer(x.ptr), x.length)
	slice.ptr = (*byte)(buf)
	slice.len = x.length
	slice.cap = x.length
	return
}

// Convert a []rune slice to a string.
func stringFromRunes(runeSlice []rune) (s _string) {
	// Count the number of characters that will be in the string.
	for _, r := range runeSlice {
		_, numBytes := encodeUTF8(r)
		s.length += numBytes
	}

	// Allocate memory for the string.
	s.ptr = (*byte)(alloc(s.length))

	// Encode runes to UTF-8 and store the resulting bytes in the string.
	index := uintptr(0)
	for _, r := range runeSlice {
		array, numBytes := encodeUTF8(r)
		for _, c := range array[:numBytes] {
			*(*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(s.ptr)) + index)) = c
			index++
		}
	}

	return
}

// Convert a string to []rune slice.
func stringToRunes(s string) []rune {
	var n = 0
	for range s {
		n++
	}
	var r = make([]rune, n)
	n = 0
	for _, e := range s {
		r[n] = e
		n++
	}
	return r
}

// Create a string from a Unicode code point.
func stringFromUnicode(x rune) _string {
	array, length := encodeUTF8(x)
	// Array will be heap allocated.
	// The heap most likely doesn't work with blocks below 4 bytes, so there's
	// no point in allocating a smaller buffer for the string here.
	return _string{ptr: (*byte)(unsafe.Pointer(&array)), length: length}
}

// Iterate over a string.
// Returns (ok, key, value).
func stringNext(s string, it *stringIterator) (bool, int, rune) {
	if len(s) <= int(it.byteindex) {
		return false, 0, 0
	}
	i := int(it.byteindex)
	r, length := decodeUTF8(s, it.byteindex)
	it.byteindex += length
	return true, i, r
}

// Convert a Unicode code point into an array of bytes and its length.
func encodeUTF8(x rune) ([4]byte, uintptr) {
	// https://stackoverflow.com/questions/6240055/manually-converting-unicode-codepoints-into-utf-8-and-utf-16
	// Note: this code can probably be optimized (in size and speed).
	switch {
	case x <= 0x7f:
		return [4]byte{byte(x), 0, 0, 0}, 1
	case x <= 0x7ff:
		b1 := 0xc0 | byte(x>>6)
		b2 := 0x80 | byte(x&0x3f)
		return [4]byte{b1, b2, 0, 0}, 2
	case x <= 0xffff:
		b1 := 0xe0 | byte(x>>12)
		b2 := 0x80 | byte((x>>6)&0x3f)
		b3 := 0x80 | byte((x>>0)&0x3f)
		return [4]byte{b1, b2, b3, 0}, 3
	case x <= 0x10ffff:
		b1 := 0xf0 | byte(x>>18)
		b2 := 0x80 | byte((x>>12)&0x3f)
		b3 := 0x80 | byte((x>>6)&0x3f)
		b4 := 0x80 | byte((x>>0)&0x3f)
		return [4]byte{b1, b2, b3, b4}, 4
	default:
		// Invalid Unicode code point.
		return [4]byte{0xef, 0xbf, 0xbd, 0}, 3
	}
}

// Decode a single UTF-8 character from a string.
//go:nobounds
func decodeUTF8(s string, index uintptr) (rune, uintptr) {
	remaining := uintptr(len(s)) - index // must be >= 1 before calling this function
	x := s[index]
	switch {
	case x&0x80 == 0x00: // 0xxxxxxx
		return rune(x), 1
	case x&0xe0 == 0xc0: // 110xxxxx
		if remaining < 2 {
			return 0xfffd, 1
		}
		return (rune(x&0x1f) << 6) | (rune(s[index+1]) & 0x3f), 2
	case x&0xf0 == 0xe0: // 1110xxxx
		if remaining < 3 {
			return 0xfffd, 1
		}
		return (rune(x&0x0f) << 12) | ((rune(s[index+1]) & 0x3f) << 6) | (rune(s[index+2]) & 0x3f), 3
	case x&0xf8 == 0xf0: // 11110xxx
		if remaining < 4 {
			return 0xfffd, 1
		}
		return (rune(x&0x07) << 18) | ((rune(s[index+1]) & 0x3f) << 12) | ((rune(s[index+2]) & 0x3f) << 6) | (rune(s[index+3]) & 0x3f), 4
	default:
		return 0xfffd, 1
	}
}

// indexByte returns the index of the first instance of c in the byte slice b,
// or -1 if c is not present in the byte slice.
//go:linkname indexByte internal/bytealg.IndexByte
func indexByte(b []byte, c byte) int {
	for i, x := range b {
		if x == c {
			return i
		}
	}
	return -1
}

// indexByteString returns the index of the first instance of c in s, or -1 if c
// is not present in s.
//go:linkname indexByteString internal/bytealg.IndexByteString
func indexByteString(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// countString copies the implementation from
// https://github.com/golang/go/blob/67f181bfd84dfd5942fe9a29d8a20c9ce5eb2fea/src/internal/bytealg/count_generic.go#L1
//go:linkname countString internal/bytealg.CountString
func countString(s string, c byte) int {
	n := 0
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			n++
		}
	}
	return n
}

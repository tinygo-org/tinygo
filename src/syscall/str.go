// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package syscall

import "errors"

// clen returns the index of the first NULL byte in n or len(n) if n contains no NULL byte.
func clen(n []byte) int {
	for i := 0; i < len(n); i++ {
		if n[i] == 0 {
			return i
		}
	}
	return len(n)
}

// BytePtrFromString returns a pointer to a NUL-terminated array of
// bytes containing the text of s. If s contains a NUL byte at any
// location, it returns (nil, [EINVAL]).
//
// For darwin || nintendoswitch || wasi || wasip1 this is already implemented in /src/syscall/syscall_libc.go:266:6
// func BytePtrFromString(s string) (*byte, error) {
// 	a, err := ByteSliceFromString(s)
// 	if err != nil {
// 		return nil, err
// 	}
// 	return &a[0], nil
// }

// copied from upstream src/internal/bytealg/indexbyte_generic.go since we cannot use the internal bytealg package
func IndexByteString(s string, c byte) int {
	for i := 0; i < len(s); i++ {
		if s[i] == c {
			return i
		}
	}
	return -1
}

// ByteSliceFromString returns a NUL-terminated slice of bytes
// containing the text of s. If s contains a NUL byte at any
// location, it returns (nil, [EINVAL]).
// https://cs.opensource.google/go/go/+/master:src/syscall/syscall.go;l=45;drc=94982a07825aec711f11c97283e99e467838d616
func ByteSliceFromString(s string) ([]byte, error) {
	if IndexByteString(s, 0) != -1 {
		return nil, errors.New("contains NUL")
	}
	a := make([]byte, len(s)+1)
	copy(a, s)
	return a, nil
}

func SlicePtrFromStrings(ss []string) ([]*byte, error) {
	n := 0
	for _, s := range ss {
		if IndexByteString(s, 0) != -1 {
			return nil, EINVAL
		}
		n += len(s) + 1 // +1 for NUL
	}
	bb := make([]*byte, len(ss)+1)
	b := make([]byte, n)
	n = 0
	for i, s := range ss {
		bb[i] = &b[n]
		copy(b[n:], s)
		n += len(s) + 1
	}
	return bb, nil
}

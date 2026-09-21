// Copyright 2010 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Package rand implements a cryptographically secure
// random number generator.
package rand

import "io"

const base32alphabet = "ABCDEFGHIJKLMNOPQRSTUVWXYZ234567"

// Reader is a global, shared instance of a cryptographically
// secure random number generator.
var Reader io.Reader

// Read is a helper function that calls Reader.Read using io.ReadFull.
// On return, n == len(b) if and only if err == nil.
func Read(b []byte) (n int, err error) {
	if Reader == nil {
		panic("no rng")
	}

	return io.ReadFull(Reader, b)
}

// Text returns a cryptographically random string using the standard RFC 4648
// base32 alphabet.
func Text() string {
	src := make([]byte, 26)
	if _, err := Read(src); err != nil {
		// TODO: Remove this check when Read terminates on errors, as in Go.
		panic("crypto/rand: failed to read random data: " + err.Error())
	}
	for i := range src {
		src[i] = base32alphabet[src[i]%32]
	}
	return string(src)
}

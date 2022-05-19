//go:build wasi
// +build wasi

// The following is copied from Go 1.18 official implementation.

// Copyright 2015 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package net

import (
	"os"
)

type fileAddr string

func (fileAddr) Network() string  { return "file+net" }
func (f fileAddr) String() string { return string(f) }

// FileListener returns a copy of the network listener corresponding
// to the open file f.
// It is the caller's responsibility to close ln when finished.
// Closing ln does not affect f, and closing f does not affect ln.
func FileListener(f *os.File) (ln Listener, err error) {
	ln, err = fileListener(f)
	if err != nil {
		err = &OpError{Op: "file", Net: "file+net", Source: nil, Addr: fileAddr(f.Name()), Err: err}
	}
	return
}

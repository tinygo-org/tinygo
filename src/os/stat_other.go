//go:build baremetal || (wasm && !wasi)

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

// Sync is a stub, not yet implemented
func (f *File) Sync() error {
	return ErrNotImplemented
}

// Stat is a stub, not yet implemented
func (f *File) Stat() (FileInfo, error) {
	return nil, ErrNotImplemented
}

// statNolog stats a file with no test logging.
func statNolog(name string) (FileInfo, error) {
	return nil, &PathError{Op: "stat", Path: name, Err: ErrNotImplemented}
}

// lstatNolog lstats a file with no test logging.
func lstatNolog(name string) (FileInfo, error) {
	return nil, &PathError{Op: "lstat", Path: name, Err: ErrNotImplemented}
}

//go:build baremetal || (tinygo.wasm && !wasip1 && !wasip2) || nintendoswitch

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

// Stat returns a FileInfo describing the file.
func (f *File) Stat() (FileInfo, error) {
	if f == nil {
		return nil, ErrInvalid
	}
	if f.handle == nil {
		return nil, &PathError{Op: "stat", Path: f.name, Err: ErrClosed}
	}
	if _, ok := f.handle.(nullFileHandle); ok {
		return nullFileInfo{}, nil
	}
	return nil, ErrNotImplemented
}

// statNolog stats a file with no test logging.
func statNolog(name string) (FileInfo, error) {
	if isNullDevice(name) {
		return nullFileInfo{}, nil
	}
	return nil, &PathError{Op: "stat", Path: name, Err: ErrNotImplemented}
}

// lstatNolog lstats a file with no test logging.
func lstatNolog(name string) (FileInfo, error) {
	if isNullDevice(name) {
		return nullFileInfo{}, nil
	}
	return nil, &PathError{Op: "lstat", Path: name, Err: ErrNotImplemented}
}

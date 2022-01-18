// +build linux,!baremetal,386 linux,!baremetal,arm,!wasi

// Functions broken by lack of seek().
// Stat is broken because it uses Time, which has a preadn function that uses seek :-(
//
// TODO: remove this file once tinygo gets syscall.Seek support on i386

// Copyright 2016 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

// Seek sets the offset for the next Read or Write on file to offset, interpreted
// according to whence: 0 means relative to the origin of the file, 1 means
// relative to the current offset, and 2 means relative to the end.
// It returns the new offset and an error, if any.
// The behavior of Seek on a file opened with O_APPEND is not specified.
//
// If f is a directory, the behavior of Seek varies by operating
// system; you can seek to the beginning of the directory on Unix-like
// operating systems, but not on Windows.
func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	return f.handle.Seek(offset, whence)
}

// Stat returns the FileInfo structure describing file.
// If there is an error, it will be of type *PathError.
func (f *File) Stat() (FileInfo, error) {
	return nil, &PathError{Op: "fstat", Path: f.name, Err: ErrNotImplemented}
}

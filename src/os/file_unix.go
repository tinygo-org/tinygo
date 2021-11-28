// +build darwin linux,!baremetal

// Portions copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"io"
	"syscall"
)

type syscallFd = int

func Pipe() (r *File, w *File, err error) {
	var p [2]int
	err = handleSyscallError(syscall.Pipe2(p[:], syscall.O_CLOEXEC))
	if err != nil {
		return
	}
	r = &File{
		handle: unixFileHandle(p[0]),
		name:   "|0",
	}
	w = &File{
		handle: unixFileHandle(p[1]),
		name:   "|1",
	}
	return
}

func tempDir() string {
	dir := Getenv("TMPDIR")
	if dir == "" {
		dir = "/tmp"
	}
	return dir
}

// ReadAt reads up to len(b) bytes from the File starting at the given absolute offset.
// It returns the number of bytes read and any error encountered, possibly io.EOF.
// At end of file, Pread returns 0, io.EOF.
// TODO: move to file_anyos once ReadAt is implemented for windows
func (f unixFileHandle) ReadAt(b []byte, offset int64) (n int, err error) {
	n, err = syscall.Pread(syscallFd(f), b, offset)
	err = handleSyscallError(err)
	if n == 0 && err == nil {
		err = io.EOF
	}
	return
}

// ignoringEINTR makes a function call and repeats it if it returns an
// EINTR error. This appears to be required even though we install all
// signal handlers with SA_RESTART: see #22838, #38033, #38836, #40846.
// Also #20400 and #36644 are issues in which a signal handler is
// installed without setting SA_RESTART. None of these are the common case,
// but there are enough of them that it seems that we can't avoid
// an EINTR loop.
func ignoringEINTR(fn func() error) error {
	for {
		err := fn()
		if err != syscall.EINTR {
			return err
		}
	}
}

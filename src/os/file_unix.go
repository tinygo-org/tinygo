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

// fixLongPath is a noop on non-Windows platforms.
func fixLongPath(path string) string {
	return path
}

func rename(oldname, newname string) error {
	// TODO: import rest of upstream tests, handle fancy cases
	err := syscall.Rename(oldname, newname)
	if err != nil {
		return &LinkError{"rename", oldname, newname, err}
	}
	return nil
}

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

// Symlink creates newname as a symbolic link to oldname.
// On Windows, a symlink to a non-existent oldname creates a file symlink;
// if oldname is later created as a directory the symlink will not work.
// If there is an error, it will be of type *LinkError.
func Symlink(oldname, newname string) error {
	e := ignoringEINTR(func() error {
		return syscall.Symlink(oldname, newname)
	})
	if e != nil {
		return &LinkError{"symlink", oldname, newname, e}
	}
	return nil
}

// Readlink returns the destination of the named symbolic link.
// If there is an error, it will be of type *PathError.
func Readlink(name string) (string, error) {
	for len := 128; ; len *= 2 {
		b := make([]byte, len)
		var (
			n int
			e error
		)
		for {
			n, e = fixCount(syscall.Readlink(name, b))
			if e != syscall.EINTR {
				break
			}
		}
		if e != nil {
			return "", &PathError{Op: "readlink", Path: name, Err: e}
		}
		if n < len {
			return string(b[0:n]), nil
		}
	}
}

// ReadAt reads up to len(b) bytes from the File starting at the given absolute offset.
// It returns the number of bytes read and any error encountered, possibly io.EOF.
// At end of file, Pread returns 0, io.EOF.
// TODO: move to file_anyos once ReadAt is implemented for windows
func (f unixFileHandle) ReadAt(b []byte, offset int64) (n int, err error) {
	n, err = syscall.Pread(syscallFd(f), b, offset)
	err = handleSyscallError(err)
	if n == 0 && len(b) > 0 && err == nil {
		err = io.EOF
	}
	return
}

// Seek wraps syscall.Seek.
func (f unixFileHandle) Seek(offset int64, whence int) (int64, error) {
	newoffset, err := syscall.Seek(syscallFd(f), offset, whence)
	return newoffset, handleSyscallError(err)
}

//go:build windows
// +build windows

// Portions copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"internal/syscall/windows"
	"syscall"
	"unicode/utf16"
)

type syscallFd = syscall.Handle

// Symlink is a stub, it is not implemented.
func Symlink(oldname, newname string) error {
	return ErrNotImplemented
}

// Readlink is a stub (for now), always returning the string it was given
func Readlink(name string) (string, error) {
	return name, nil
}

func rename(oldname, newname string) error {
	e := windows.Rename(fixLongPath(oldname), fixLongPath(newname))
	if e != nil {
		return &LinkError{"rename", oldname, newname, e}
	}
	return nil
}

func Pipe() (r *File, w *File, err error) {
	var p [2]syscall.Handle
	e := handleSyscallError(syscall.Pipe(p[:]))
	if e != nil {
		return nil, nil, err
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
	n := uint32(syscall.MAX_PATH)
	for {
		b := make([]uint16, n)
		n, _ = syscall.GetTempPath(uint32(len(b)), &b[0])
		if n > uint32(len(b)) {
			continue
		}
		if n == 3 && b[1] == ':' && b[2] == '\\' {
			// Do nothing for path, like C:\.
		} else if n > 0 && b[n-1] == '\\' {
			// Otherwise remove terminating \.
			n--
		}
		return string(utf16.Decode(b[:n]))
	}
}

// ReadAt reads up to len(b) bytes from the File starting at the given absolute offset.
// It returns the number of bytes read and any error encountered, possibly io.EOF.
// At end of file, Pread returns 0, io.EOF.
// TODO: move to file_anyos once ReadAt is implemented for windows
func (f unixFileHandle) ReadAt(b []byte, offset int64) (n int, err error) {
	return -1, ErrNotImplemented
}

// Seek wraps syscall.Seek.
func (f unixFileHandle) Seek(offset int64, whence int) (int64, error) {
	newoffset, err := syscall.Seek(syscallFd(f), offset, whence)
	return newoffset, handleSyscallError(err)
}

// Seek sets the offset for the next Read or Write on file to offset, interpreted
// according to whence: 0 means relative to the origin of the file, 1 means
// relative to the current offset, and 2 means relative to the end.
// It returns the new offset and an error, if any.
// The behavior of Seek on a file opened with O_APPEND is not specified.
//
// If f is a directory, the behavior of Seek varies by operating
// system; you can seek to the beginning of the directory on Unix-like
// operating systems, but not on Windows.
// TODO: move this back to file.go once syscall.seek is implemented on 386 and arm.
func (f *File) Seek(offset int64, whence int) (ret int64, err error) {
	return f.handle.Seek(offset, whence)
}

// isWindowsNulName reports whether name is os.DevNull ('NUL') on Windows.
// True is returned if name is 'NUL' whatever the case.
func isWindowsNulName(name string) bool {
	if len(name) != 3 {
		return false
	}
	if name[0] != 'n' && name[0] != 'N' {
		return false
	}
	if name[1] != 'u' && name[1] != 'U' {
		return false
	}
	if name[2] != 'l' && name[2] != 'L' {
		return false
	}
	return true
}

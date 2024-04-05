//go:build windows

// Portions copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"internal/syscall/windows"
	"syscall"
	"unicode/utf16"
)

const DevNull = "NUL"

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

type file struct {
	handle     FileHandle
	name       string
	appendMode bool
}

func (f *file) close() error {
	return f.handle.Close()
}

func NewFile(fd uintptr, name string) *File {
	return &File{&file{handle: unixFileHandle(fd), name: name}}
}

func Pipe() (r *File, w *File, err error) {
	var p [2]syscall.Handle
	e := handleSyscallError(syscall.Pipe(p[:]))
	if e != nil {
		return nil, nil, err
	}
	r = NewFile(uintptr(p[0]), "|0")
	w = NewFile(uintptr(p[1]), "|1")
	return
}

func (f *unixFileHandle) Truncate(size int64) error {
	return ErrNotImplemented
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

// WriteAt writes len(b) bytes to the File starting at byte offset off.
// It returns the number of bytes written and an error, if any.
// WriteAt returns a non-nil error when n != len(b).
//
// If file was opened with the O_APPEND flag, WriteAt returns an error.
//
// TODO: move to file_anyos once WriteAt is implemented for windows.
func (f unixFileHandle) WriteAt(b []byte, offset int64) (n int, err error) {
	return -1, ErrNotImplemented
}

// Seek wraps syscall.Seek.
func (f unixFileHandle) Seek(offset int64, whence int) (int64, error) {
	newoffset, err := syscall.Seek(syscallFd(f), offset, whence)
	return newoffset, handleSyscallError(err)
}

func (f unixFileHandle) Sync() error {
	return ErrNotImplemented
}

func (f *File) Truncate(size int64) error {
	var err error
	if f.handle == nil {
		err = ErrClosed
	} else {
		err = ErrNotImplemented
	}
	return &PathError{Op: "truncate", Path: f.name, Err: err}
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

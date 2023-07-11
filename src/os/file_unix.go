//go:build darwin || (linux && !baremetal)

// target wasi sets GOOS=linux and thus the +linux build tag,
// even though it doesn't show up in "tinygo info target -wasi"

// Portions copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

package os

import (
	"io"
	"syscall"
)

const DevNull = "/dev/null"

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

// file is the real representation of *File.
// The extra level of indirection ensures that no clients of os
// can overwrite this data, which could cause the finalizer
// to close the wrong file descriptor.
type file struct {
	handle     FileHandle
	name       string
	dirinfo    *dirInfo // nil unless directory being read
	appendMode bool
}

func (f *file) close() (err error) {
	if f.dirinfo != nil {
		f.dirinfo.close()
		f.dirinfo = nil
	}
	return f.handle.Close()
}

func NewFile(fd uintptr, name string) *File {
	return &File{&file{handle: unixFileHandle(fd), name: name}}
}

func Pipe() (r *File, w *File, err error) {
	var p [2]int
	err = handleSyscallError(syscall.Pipe2(p[:], syscall.O_CLOEXEC))
	if err != nil {
		return
	}
	r = NewFile(uintptr(p[0]), "|0")
	w = NewFile(uintptr(p[1]), "|1")
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

// WriteAt writes len(b) bytes to the File starting at byte offset off.
// It returns the number of bytes written and an error, if any.
// WriteAt returns a non-nil error when n != len(b).
//
// If file was opened with the O_APPEND flag, WriteAt returns an error.
//
// TODO: move to file_anyos once WriteAt is implemented for windows.
func (f unixFileHandle) WriteAt(b []byte, offset int64) (int, error) {
	n, err := syscall.Pwrite(syscallFd(f), b, offset)
	return n, handleSyscallError(err)
}

// Seek wraps syscall.Seek.
func (f unixFileHandle) Seek(offset int64, whence int) (int64, error) {
	newoffset, err := syscall.Seek(syscallFd(f), offset, whence)
	return newoffset, handleSyscallError(err)
}

func (f unixFileHandle) Sync() error {
	err := syscall.Fsync(syscallFd(f))
	return handleSyscallError(err)
}

type unixDirent struct {
	parent string
	name   string
	typ    FileMode
	info   FileInfo
}

func (d *unixDirent) Name() string   { return d.name }
func (d *unixDirent) IsDir() bool    { return d.typ.IsDir() }
func (d *unixDirent) Type() FileMode { return d.typ }

func (d *unixDirent) Info() (FileInfo, error) {
	if d.info != nil {
		return d.info, nil
	}
	return lstat(d.parent + "/" + d.name)
}

func newUnixDirent(parent, name string, typ FileMode) (DirEntry, error) {
	ude := &unixDirent{
		parent: parent,
		name:   name,
		typ:    typ,
	}
	if typ != ^FileMode(0) && !testingForceReadDirLstat {
		return ude, nil
	}

	info, err := lstat(parent + "/" + name)
	if err != nil {
		return nil, err
	}

	ude.typ = info.Mode().Type()
	ude.info = info
	return ude, nil
}

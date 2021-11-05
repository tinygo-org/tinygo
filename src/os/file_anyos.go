// +build !baremetal,!js

package os

import (
	"io"
	"syscall"
)

func init() {
	// Mount the host filesystem at the root directory. This is what most
	// programs will be expecting.
	Mount("/", unixFilesystem{})
}

// Stdin, Stdout, and Stderr are open Files pointing to the standard input,
// standard output, and standard error file descriptors.
var (
	Stdin  = &File{unixFileHandle(syscall.Stdin), "/dev/stdin"}
	Stdout = &File{unixFileHandle(syscall.Stdout), "/dev/stdout"}
	Stderr = &File{unixFileHandle(syscall.Stderr), "/dev/stderr"}
)

// isOS indicates whether we're running on a real operating system with
// filesystem support.
const isOS = true

// unixFilesystem is an empty handle for a Unix/Linux filesystem. All operations
// are relative to the current working directory.
type unixFilesystem struct {
}

func (fs unixFilesystem) Mkdir(path string, perm FileMode) error {
	return handleSyscallError(syscall.Mkdir(path, uint32(perm)))
}

func (fs unixFilesystem) Remove(path string) error {
	return handleSyscallError(syscall.Unlink(path))
}

func (fs unixFilesystem) OpenFile(path string, flag int, perm FileMode) (FileHandle, error) {
	// Map os package flags to syscall flags.
	syscallFlag := 0
	if flag&O_RDONLY != 0 {
		syscallFlag |= syscall.O_RDONLY
	}
	if flag&O_WRONLY != 0 {
		syscallFlag |= syscall.O_WRONLY
	}
	if flag&O_RDWR != 0 {
		syscallFlag |= syscall.O_RDWR
	}
	if flag&O_APPEND != 0 {
		syscallFlag |= syscall.O_APPEND
	}
	if flag&O_CREATE != 0 {
		syscallFlag |= syscall.O_CREAT
	}
	if flag&O_EXCL != 0 {
		syscallFlag |= syscall.O_EXCL
	}
	if flag&O_SYNC != 0 {
		syscallFlag |= syscall.O_SYNC
	}
	if flag&O_TRUNC != 0 {
		syscallFlag |= syscall.O_TRUNC
	}
	fp, err := syscall.Open(path, syscallFlag, uint32(perm))
	return unixFileHandle(fp), handleSyscallError(err)
}

// unixFileHandle is a Unix file pointer with associated methods that implement
// the FileHandle interface.
type unixFileHandle uintptr

// Read reads up to len(b) bytes from the File. It returns the number of bytes
// read and any error encountered. At end of file, Read returns 0, io.EOF.
func (f unixFileHandle) Read(b []byte) (n int, err error) {
	n, err = syscall.Read(syscallFd(f), b)
	err = handleSyscallError(err)
	if n == 0 && err == nil {
		err = io.EOF
	}
	return
}

// Write writes len(b) bytes to the File. It returns the number of bytes written
// and an error, if any. Write returns a non-nil error when n != len(b).
func (f unixFileHandle) Write(b []byte) (n int, err error) {
	n, err = syscall.Write(syscallFd(f), b)
	err = handleSyscallError(err)
	return
}

// Close closes the File, rendering it unusable for I/O.
func (f unixFileHandle) Close() error {
	return handleSyscallError(syscall.Close(syscallFd(f)))
}

// handleSyscallError converts syscall errors into regular os package errors.
// The err parameter must be either nil or of type syscall.Errno.
func handleSyscallError(err error) error {
	if err == nil {
		return nil
	}
	switch err.(syscall.Errno) {
	case syscall.EEXIST:
		return ErrExist
	case syscall.ENOENT:
		return ErrNotExist
	default:
		return err
	}
}

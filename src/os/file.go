// Package os implements a subset of the Go "os" package. See
// https://godoc.org/os for details.
//
// Note that the current implementation is blocking. This limitation should be
// removed in a future version.
package os

import (
	"errors"
)

// Portable analogs of some common system call errors.
var (
	ErrUnsupported = errors.New("operation not supported")
)

// Stdin, Stdout, and Stderr are open Files pointing to the standard input,
// standard output, and standard error file descriptors.
var (
	Stdin  = &File{0, "/dev/stdin"}
	Stdout = &File{1, "/dev/stdout"}
	Stderr = &File{2, "/dev/stderr"}
)

// File represents an open file descriptor.
type File struct {
	fd   uintptr
	name string
}

// NewFile returns a new File with the given file descriptor and name.
func NewFile(fd uintptr, name string) *File {
	return &File{fd, name}
}

// Fd returns the integer Unix file descriptor referencing the open file. The
// file descriptor is valid only until f.Close is called.
func (f *File) Fd() uintptr {
	return f.fd
}

// +build baremetal wasm,!wasi

package os

import (
	_ "unsafe"
)

// Stdin, Stdout, and Stderr are open Files pointing to the standard input,
// standard output, and standard error file descriptors.
var (
	Stdin  = &File{stdioFileHandle(0), "/dev/stdin"}
	Stdout = &File{stdioFileHandle(1), "/dev/stdout"}
	Stderr = &File{stdioFileHandle(2), "/dev/stderr"}
)

// isOS indicates whether we're running on a real operating system with
// filesystem support.
const isOS = false

// stdioFileHandle represents one of stdin, stdout, or stderr depending on the
// number. It implements the FileHandle interface.
type stdioFileHandle uint8

// Read is unsupported on this system.
func (f stdioFileHandle) Read(b []byte) (n int, err error) {
	return 0, ErrUnsupported
}

func (f stdioFileHandle) ReadAt(b []byte, off int64) (n int, err error) {
	return 0, ErrNotImplemented
}

// Write writes len(b) bytes to the output. It returns the number of bytes
// written or an error if this file is not stdout or stderr.
func (f stdioFileHandle) Write(b []byte) (n int, err error) {
	switch f {
	case 1, 2: // stdout, stderr
		for _, c := range b {
			putchar(c)
		}
		return len(b), nil
	default:
		return 0, ErrUnsupported
	}
}

// Close is unsupported on this system.
func (f stdioFileHandle) Close() error {
	return ErrUnsupported
}

// Seek wraps syscall.Seek.
func (f stdioFileHandle) Seek(offset int64, whence int) (int64, error) {
	return -1, ErrUnsupported
}

//go:linkname putchar runtime.putchar
func putchar(c byte)

func Pipe() (r *File, w *File, err error) {
	return nil, nil, ErrNotImplemented
}

func Readlink(name string) (string, error) {
	return "", ErrNotImplemented
}

func tempDir() string {
	return "/tmp"
}

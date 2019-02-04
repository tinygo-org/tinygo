// +build wasm

package os

import (
	_ "unsafe"
)

// Read is unsupported on this system.
func (f *File) Read(b []byte) (n int, err error) {
	return 0, ErrUnsupported
}

// Write writes len(b) bytes to the output. It returns the number of bytes
// written or an error if this file is not stdout or stderr.
func (f *File) Write(b []byte) (n int, err error) {
	switch f.fd {
	case Stdout.fd, Stderr.fd:
		for _, c := range b {
			putchar(c)
		}
		return len(b), nil
	default:
		return 0, ErrUnsupported
	}
}

// Close is unsupported on this system.
func (f *File) Close() error {
	return ErrUnsupported
}

//go:linkname putchar runtime.putchar
func putchar(c byte)

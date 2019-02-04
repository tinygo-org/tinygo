// +build linux

package os

import (
	"syscall"
)

// Read reads up to len(b) bytes from the File. It returns the number of bytes
// read and any error encountered. At end of file, Read returns 0, io.EOF.
func (f *File) Read(b []byte) (n int, err error) {
	return syscall.Read(int(f.fd), b)
}

// Write writes len(b) bytes to the File. It returns the number of bytes written
// and an error, if any. Write returns a non-nil error when n != len(b).
func (f *File) Write(b []byte) (n int, err error) {
	return syscall.Write(int(f.fd), b)
}

// Close closes the File, rendering it unusable for I/O.
func (f *File) Close() error {
	return syscall.Close(int(f.fd))
}

// +build darwin linux,!baremetal

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

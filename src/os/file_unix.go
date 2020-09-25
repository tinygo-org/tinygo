// +build darwin linux,!baremetal

package os

import "syscall"

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

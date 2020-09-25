// +build windows

package os

import "syscall"

type syscallFd = syscall.Handle

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

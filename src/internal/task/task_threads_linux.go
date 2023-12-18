package task

import (
	"strconv"
	"unsafe"
)

//go:extern errno
var libcErrno uintptr

//export pthread_create
func pthread_create(pthread uintptr, arg2 uintptr, fn uintptr, param uintptr) int

type Errno uintptr

func (e Errno) Error() string {
	return "system error: " + strconv.Itoa(int(e))
}

func createThread(fn uintptr, param uintptr, stackSize uintptr) (uintptr, error) {
	var pthread uintptr
	if pthread_create(uintptr(unsafe.Pointer(&pthread)), 0, fn, param) < 0 {
		return pthread, Errno(libcErrno)
	}
	return pthread, nil
}

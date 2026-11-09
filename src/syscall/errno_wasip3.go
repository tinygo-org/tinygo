//go:build wasip3

package syscall

import "unsafe"

// With threads, errno of wasi-libc is thread-local. Get it via the function.
//
//export __errno_location
func libc_errno_location() *int32

func getLibcErrno() Errno {
	return Errno(*(*int32)(unsafe.Pointer(libc_errno_location())))
}

func setLibcErrno(e Errno) {
	*libc_errno_location() = int32(e)
}

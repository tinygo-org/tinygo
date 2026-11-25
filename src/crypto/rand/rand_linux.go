//go:build !baremetal && !wasip2 && !wasm_unknown && !nintendoswitch

// This implementation of crypto/rand uses the getrandom libc function to
// generate random numbers.

package rand

import (
	"syscall"
	"unsafe"
)

func init() {
	Reader = &reader{}
}

type reader struct {
}

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}
	for {
		// getrandom can return fewer bytes than asked for, or fail with EINTR.
		// See https://man7.org/linux/man-pages/man2/getrandom.2.html
		n := libc_getrandom(unsafe.Pointer(unsafe.SliceData(b)), uint(len(b)), 0)
		if n >= 0 {
			return n, nil
		}
		if errno := syscall.Errno(*libc_errno_location()); errno != syscall.EINTR {
			return 0, errno
		}
	}
}

// ssize_t getrandom(void *buf, size_t buflen, unsigned int flags);
//
//export getrandom
func libc_getrandom(buf unsafe.Pointer, buflen uint, flags uint32) int

// int *__errno_location(void);
//
//export __errno_location
func libc_errno_location() *int32

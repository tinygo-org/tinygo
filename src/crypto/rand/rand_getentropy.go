// +build darwin freebsd wasi

// This implementation of crypto/rand uses the getentropy system call (available
// on both MacOS and WASI) to generate random numbers.

package rand

import (
	"errors"
	"unsafe"
)

var errReadFailed = errors.New("rand: could not read random bytes")

func init() {
	Reader = &reader{}
}

type reader struct {
}

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) != 0 {
		if len(b) > 256 {
			b = b[:256]
		}
		result := libc_getentropy(unsafe.Pointer(&b[0]), len(b))
		if result < 0 {
			// Maybe we should return a syscall.Errno here?
			return 0, errReadFailed
		}
	}
	return len(b), nil
}

// int getentropy(void *buf, size_t buflen);
//export getentropy
func libc_getentropy(buf unsafe.Pointer, buflen int) int

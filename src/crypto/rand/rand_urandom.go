//go:build linux && !baremetal && !wasi

// This implementation of crypto/rand uses the /dev/urandom pseudo-file to
// generate random numbers.
// TODO: convert to the getentropy or getrandom libc function on Linux once it
// is more widely supported.

package rand

import (
	"syscall"
)

func init() {
	Reader = &reader{}
}

type reader struct {
	fd int
}

func (r *reader) Read(b []byte) (n int, err error) {
	if len(b) == 0 {
		return
	}

	// Open /dev/urandom first if needed.
	if r.fd == 0 {
		fd, err := syscall.Open("/dev/urandom", syscall.O_RDONLY, 0)
		if err != nil {
			return 0, err
		}
		r.fd = fd
	}

	// Read from the file.
	return syscall.Read(r.fd, b)
}

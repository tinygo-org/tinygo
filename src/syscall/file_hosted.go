//go:build !baremetal && !wasm && !wasm_freestanding
// +build !baremetal,!wasm,!wasm_freestanding

// This file assumes there is a libc available that runs on a real operating
// system.

package syscall

const pathMax = 1024

func Getwd() (string, error) {
	var buf [pathMax]byte
	s := libc_getcwd(&buf[0], uint(len(buf)))
	if s == nil {
		return "", getErrno()
	}
	n := clen(buf[:])
	if n < 1 {
		return "", EINVAL
	}
	return string(buf[:n]), nil
}

// char *getcwd(char *buf, size_t size)
//
//export getcwd
func libc_getcwd(buf *byte, size uint) *byte

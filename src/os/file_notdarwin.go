//go:build (linux && !baremetal && !wasm_unknown) || wasip1 || wasip2

package os

import "syscall"

func pipe(p []int) error {
	return syscall.Pipe2(p, syscall.O_CLOEXEC)
}

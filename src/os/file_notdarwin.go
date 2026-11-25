//go:build (linux && !baremetal && !wasm_unknown) || tinygo.wasi

package os

import "syscall"

func pipe(p []int) error {
	return syscall.Pipe2(p, syscall.O_CLOEXEC)
}

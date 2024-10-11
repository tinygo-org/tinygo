//go:build linux && !baremetal && !darwin && !tinygo.wasm && aarch64

// arm64 does not have a fork syscall, so ignore it for now
// TODO: add support for arm64 with clone or use musl implementation

package os

import (
	"errors"
)

func fork() (pid int, err error) {
	return 0, errors.New("fork not supported on aarch64")
}

func execve(pathname string, argv []string, envv []string) (err error) {
	return 0, errors.New("execve not supported on aarch64")
}

//go:build linux && !baremetal && !darwin && !tinygo.wasm && !arm64

// arm64 does not have a fork syscall, so ignore it for now
// TODO: add support for arm64 with clone or use musl implementation

package os

import (
	"syscall"
	"unsafe"
)

func fork() (pid int, err error) {
	ret, _, err := syscall.Syscall(syscall.SYS_FORK, 0, 0, 0)
	if int(ret) != 0 {
		errno := err.(syscall.Errno)
		return 0, errno
	}
	return int(ret), nil
}

// the golang standard library does not expose interfaces for execve and fork, so we define them here the same way via the libc wrapper
func execve(pathname string, argv []string, envv []string) error {
	argv0 := cstring(pathname)

	// transform argv and envv into the format expected by execve
	argv1 := make([]*byte, len(argv)+1)
	for i, arg := range argv {
		argv1[i] = &cstring(arg)[0]
	}
	argv1[len(argv)] = nil

	env1 := make([]*byte, len(envv)+1)
	for i, env := range envv {
		env1[i] = &cstring(env)[0]
	}
	env1[len(envv)] = nil

	ret, _, err := syscall.Syscall(syscall.SYS_EXECVE, uintptr(unsafe.Pointer(&argv0[0])), uintptr(unsafe.Pointer(&argv1[0])), uintptr(unsafe.Pointer(&env1[0])))
	if int(ret) != 0 {
		// errno := err.(syscall.Errno)
		// return errno
		return err
	}

	return nil
}

func cstring(s string) []byte {
	data := make([]byte, len(s)+1)
	copy(data, s)
	// final byte should be zero from the initial allocation
	return data
}

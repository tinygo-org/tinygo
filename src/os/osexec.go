//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"syscall"
	"unsafe"
)

func fork() (pid int32, err error) {
	pid = libc_fork()
	if pid != 0 {
		if errno := *libc_errno(); errno != 0 {
			err = syscall.Errno(*libc_errno())
		}
	}
	return
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

//export fork
func libc_fork() int32

// Internal musl function to get the C errno pointer.
//
//export __errno_location
func libc_errno() *int32

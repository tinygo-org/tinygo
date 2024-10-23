//go:build linux || unix
// +build linux unix

package syscall

import "errors"

func Exec(argv0 string, argv []string, envv []string) (err error)

// The two SockaddrInet* structs have been copied from the Go source tree.

type SockaddrInet4 struct {
	Port int
	Addr [4]byte
	raw  RawSockaddrInet4
}

type SockaddrInet6 struct {
	Port   int
	ZoneId uint32
	Addr   [16]byte
	raw    RawSockaddrInet6
}

func Fork() (err error) {
	fail := int(libc_fork())
	if fail < 0 {
		// TODO: parse the syscall return codes
		return errors.New("fork failed")
	}
	return
}

// the golang standard library does not expose interfaces for execve and fork, so we define them here the same way via the libc wrapper
func Execve(pathname string, argv []string, envv []string) (err error) {
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

	fail := int(libc_execve(&argv0[0], &argv1[0], &env1[0]))
	if fail < 0 {
		// TODO: parse the syscall return codes
		return errors.New("fork failed")
	}
	return
}

func cstring(s string) []byte {
	data := make([]byte, len(s)+1)
	copy(data, s)
	// final byte should be zero from the initial allocation
	return data
}

// pid_t fork(void);
//
//export fork
func libc_fork() int32

// int execve(const char *filename, char *const argv[], char *const envp[]);
//
//export execve
func libc_execve(filename *byte, argv **byte, envp **byte) int

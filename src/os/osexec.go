//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"syscall"
	"unsafe"
)

// spawn starts path with argv and envv and returns the pid of the new process.
func spawn(path string, argv, envv []string) (int, error) {
	pathp := cstring(path)
	argvp := cstrings(argv)
	envp := cstrings(envv)
	var pid int32
	if errno := libc_posix_spawn(&pid, &pathp[0], nil, nil, &argvp[0], &envp[0]); errno != 0 {
		return 0, syscall.Errno(errno)
	}
	return int(pid), nil
}

func cstring(s string) []byte {
	data := make([]byte, len(s)+1)
	copy(data, s)
	// final byte should be zero from the initial allocation
	return data
}

// cstrings returns ss as a nil terminated array of C strings.
func cstrings(ss []string) []*byte {
	p := make([]*byte, len(ss)+1)
	for i, s := range ss {
		p[i] = &cstring(s)[0]
	}
	return p
}

// posix_spawn returns an error number instead of setting errno, see
// https://pubs.opengroup.org/onlinepubs/9799919799/functions/posix_spawn.html
//
//export posix_spawn
func libc_posix_spawn(pid *int32, path *byte, fileActions, attrp unsafe.Pointer, argv, envp **byte) int32

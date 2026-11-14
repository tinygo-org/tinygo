//go:build linux && !baremetal && !tinygo.wasm && !nintendoswitch

package os

import (
	"syscall"
	"unsafe"
)

// spawn starts path with argv and the attributes in attr and returns the pid of the new process.
func spawn(path string, argv []string, attr *ProcAttr) (int, error) {
	var fa spawnFileActions
	if errno := libc_posix_spawn_file_actions_init(&fa); errno != 0 {
		return 0, syscall.Errno(errno)
	}
	defer libc_posix_spawn_file_actions_destroy(&fa)

	if attr.Dir != "" {
		dirp := cstring(attr.Dir)
		if errno := libc_posix_spawn_file_actions_addchdir_np(&fa, &dirp[0]); errno != 0 {
			return 0, syscall.Errno(errno)
		}
	}

	pathp := cstring(path)
	argvp := cstrings(argv)
	envp := cstrings(attr.Env)
	var pid int32
	if errno := libc_posix_spawn(&pid, &pathp[0], &fa, nil, &argvp[0], &envp[0]); errno != 0 {
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

// spawnFileActions has the layout of posix_spawn_file_actions_t in musl and glibc.
type spawnFileActions struct {
	_       [2]int32
	actions unsafe.Pointer
	_       [16]int32
}

// The posix_spawn functions return an error number instead of setting errno, see
// https://pubs.opengroup.org/onlinepubs/9799919799/functions/posix_spawn.html
//
//export posix_spawn
func libc_posix_spawn(pid *int32, path *byte, fileActions *spawnFileActions, attrp unsafe.Pointer, argv, envp **byte) int32

//export posix_spawn_file_actions_init
func libc_posix_spawn_file_actions_init(fa *spawnFileActions) int32

//export posix_spawn_file_actions_destroy
func libc_posix_spawn_file_actions_destroy(fa *spawnFileActions) int32

// posix_spawn_file_actions_addchdir_np is in musl 1.1.24 and glibc 2.29.
//
//export posix_spawn_file_actions_addchdir_np
func libc_posix_spawn_file_actions_addchdir_np(fa *spawnFileActions, path *byte) int32

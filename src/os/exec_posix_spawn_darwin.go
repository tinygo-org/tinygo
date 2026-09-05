//go:build darwin

package os

import (
	"syscall"
	_ "unsafe"
)

var spawnDevNull = [...]byte{'/', 'd', 'e', 'v', '/', 'n', 'u', 'l', 'l', 0}

func addSpawnClose(fa *spawnFileActions, fd int32) error {
	// Darwin rejects close actions on unopened descriptors.
	// See posix_spawn(2), ERRORS, EBADF.
	if errno := posix_spawn_file_actions_addopen(fa, fd, &spawnDevNull[0], syscall.O_RDONLY, 0); errno != 0 {
		return syscall.Errno(errno)
	}
	if errno := posix_spawn_file_actions_addclose(fa, fd); errno != 0 {
		return syscall.Errno(errno)
	}
	return nil
}

//go:linkname posix_spawn_file_actions_addopen posix_spawn_file_actions_addopen
func posix_spawn_file_actions_addopen(fa *spawnFileActions, fd int32, path *byte, flags int32, mode uint16) int32

// Darwin <spawn.h> declares both POSIX objects as opaque pointers, so the
// object a caller allocates is one pointer wide and libc allocates the rest.
// The type is uintptr because libc stores memory there that is not Go memory.
type spawnFileActions uintptr

type spawnAttr uintptr

// The sigset_t of Darwin is a 32-bit mask. The zero value is the empty set.
type sigset uint32

// checkSysProcAttr reports whether the SysProcAttr asks for something that
// posix_spawn cannot do. Darwin declares only the common fields plus Setpgid
// and Pgid.
func checkSysProcAttr(sys *syscall.SysProcAttr) error {
	return checkSysProcAttrCommon(sys)
}

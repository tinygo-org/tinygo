//go:build darwin

package os

import "syscall"

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

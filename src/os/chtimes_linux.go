//go:build !baremetal && !nintendoswitch && !wasip1 && !wasm_unknown && !wasip2

package os

import (
	"syscall"
	"time"
)

// See https://github.com/torvalds/linux/blob/master/include/linux/stat.h
const _UTIME_OMIT = (1 << 30) - 2

// Chtimes changes the access and modification times of the named file.
func Chtimes(name string, atime time.Time, mtime time.Time) error {
	var utimes [2]syscall.Timespec
	set := func(i int, t time.Time) {
		if t.IsZero() {
			utimes[i] = syscall.Timespec{Sec: _UTIME_OMIT, Nsec: _UTIME_OMIT}
		} else {
			utimes[i] = syscall.NsecToTimespec(t.UnixNano())
		}
	}
	set(0, atime)
	set(1, mtime)
	if e := syscall.UtimesNano(fixLongPath(name), utimes[0:]); e != nil {
		return &PathError{Op: "chtimes", Path: name, Err: e}
	}
	return nil
}

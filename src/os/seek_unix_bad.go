//go:build (linux && !baremetal && 386) || (linux && !baremetal && arm && !wasi)

package os

import (
	"syscall"
)

// On linux, we use upstream's syscall package.
// But we do not yet implement Go Assembly, so we don't see a few functions written in assembly there.
// In particular, on i386 and arm, the function syscall.seek is missing, breaking syscall.Seek.
// This in turn causes os.(*File).Seek, time, io/fs, and path/filepath to fail to link.
//
// To temporarily let all the above at least link, provide a stub for syscall.seek.
// This belongs in syscall, but on linux, we use upstream's syscall.
// Remove once we support Go Assembly.
// TODO: make this a non-stub, and thus fix the whole problem?

//export syscall.seek
func seek(fd int, offset int64, whence int) (newoffset int64, err syscall.Errno) {
	return 0, syscall.ENOTSUP
}

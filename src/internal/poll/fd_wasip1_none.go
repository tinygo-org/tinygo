//go:build wasip1 && !scheduler.tasks && !scheduler.asyncify

package poll

import (
	"syscall"
	"time"
)

//go:linkname runtime_netpoll_wait runtime.runtime_netpoll_wait
func runtime_netpoll_wait(fd uint32, mode uint8, timeout uint64) uint32

func wait(fd int, mode uint8) error {
	for {
		switch errno := runtime_netpoll_wait(uint32(fd), mode, 0); errno {
		case 0:
			return nil
		case wasiErrnoIntr:
			continue
		default:
			return syscall.Errno(errno)
		}
	}
}

func (fd *FD) parkUntil(mode uint8, deadline time.Time) error {
	d := time.Until(deadline)
	if d <= 0 {
		return ErrDeadlineExceeded
	}
	errno := runtime_netpoll_wait(uint32(fd.Sysfd), mode, uint64(d))
	if errno != 0 && errno != wasiErrnoIntr {
		return syscall.Errno(errno)
	}
	return nil
}

//go:build wasip1 && (scheduler.tasks || scheduler.asyncify)

package syscall

import (
	"internal/task"
	"unsafe"
)

// pollMode constants must mirror runtime/netpoll_wasip1.go's pollRead/
// pollWrite. Keep the two definitions in sync.
const (
	pollModeRead  uint8 = 1
	pollModeWrite uint8 = 2
)

//go:linkname runtime_netpoll_addwait runtime.runtime_netpoll_addwait
func runtime_netpoll_addwait(fd uint32, mode uint8) uintptr

//go:linkname runtime_netpoll_done runtime.runtime_netpoll_done
func runtime_netpoll_done(pd uintptr)

// readWritePark is the shared park-on-EAGAIN body for Read, Write, Pread,
// Pwrite. The do() callback performs the underlying libc syscall and
// returns its result; on EAGAIN we register an FD wait, suspend the
// goroutine until the cooperative scheduler's pollIO wakes us, then
// retry. EINTR retries immediately without parking.
func Write(fd int, p []byte) (n int, err error) {
	for {
		n = libc_write(int32(fd), unsafe.SliceData(p), uint(len(p)))
		if n >= 0 {
			return
		}
		switch e := getErrno(); e {
		case EAGAIN:
			wait(fd, pollModeWrite)
		case EINTR:
			// retry
		default:
			err = e
			return
		}
	}
}

func Read(fd int, p []byte) (n int, err error) {
	for {
		n = libc_read(int32(fd), unsafe.SliceData(p), uint(len(p)))
		if n >= 0 {
			return
		}
		switch e := getErrno(); e {
		case EAGAIN:
			wait(fd, pollModeRead)
		case EINTR:
			// retry
		default:
			err = e
			return
		}
	}
}

func Pread(fd int, p []byte, offset int64) (n int, err error) {
	for {
		n = libc_pread(int32(fd), unsafe.SliceData(p), uint(len(p)), offset)
		if n >= 0 {
			return
		}
		switch e := getErrno(); e {
		case EAGAIN:
			wait(fd, pollModeRead)
		case EINTR:
			// retry
		default:
			err = e
			return
		}
	}
}

func Pwrite(fd int, p []byte, offset int64) (n int, err error) {
	for {
		n = libc_pwrite(int32(fd), unsafe.SliceData(p), uint(len(p)), offset)
		if n >= 0 {
			return
		}
		switch e := getErrno(); e {
		case EAGAIN:
			wait(fd, pollModeWrite)
		case EINTR:
			// retry
		default:
			err = e
			return
		}
	}
}

// wait parks the current goroutine until the given FD is ready for the
// requested I/O direction, then deregisters it from the poll registry.
func wait(fd int, mode uint8) {
	pd := runtime_netpoll_addwait(uint32(fd), mode)
	task.Pause()
	runtime_netpoll_done(pd)
}

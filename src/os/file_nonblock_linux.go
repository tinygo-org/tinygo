//go:build linux && !baremetal && !wasm_unknown && !nintendoswitch && !wasip1 && !wasip2 && !tinygo.wasm

package os

import (
	"sync"
	"syscall"
	"time"
	"unsafe"
)

// nbState tracks goroutines waiting for a non-blocking fd to become ready.
// Go registers such fds (os.NewFile on an O_NONBLOCK fd) with its poller so
// Read and Write never return EAGAIN. TinyGo waits in ppoll(2) instead.
type nbState struct {
	mu      sync.Mutex
	waiters int
	closing bool
}

// wouldBlock reports whether err means a non-blocking fd is not ready.
func (f *file) wouldBlock(h FileHandle, err error) bool {
	_, ok := h.(unixFileHandle)
	return ok && err == syscall.EAGAIN
}

// waitIO parks until the fd is readable (or writable when write) and returns
// ErrClosed if the file is closed meanwhile.
func (f *file) waitIO(h FileHandle, write bool) error {
	fd, ok := h.(unixFileHandle)
	if !ok {
		return syscall.EAGAIN
	}
	s := &f.nb
	s.mu.Lock()
	if s.closing {
		s.mu.Unlock()
		return ErrClosed
	}
	s.waiters++
	s.mu.Unlock()
	defer func() {
		s.mu.Lock()
		s.waiters--
		last := s.closing && s.waiters == 0
		s.mu.Unlock()
		if last {
			fd.Close()
		}
	}()

	pfd := struct {
		fd      int32
		events  int16
		revents int16
	}{fd: int32(fd), events: pollIn}
	if write {
		pfd.events = pollOut
	}
	for {
		// Wake periodically so a Close from another goroutine is noticed.
		ts := syscall.NsecToTimespec(int64(50 * time.Millisecond))
		if !pollBlocks {
			ts = syscall.Timespec{}
		}
		n, _, e := syscall.Syscall6(syscall.SYS_PPOLL, uintptr(unsafe.Pointer(&pfd)), 1, uintptr(unsafe.Pointer(&ts)), 0, 0, 0)
		s.mu.Lock()
		closing := s.closing
		s.mu.Unlock()
		if closing {
			return ErrClosed
		}
		if n > 0 || (e != 0 && e != syscall.EINTR) {
			return nil // ready, or let the retried call report the error
		}
		if !pollBlocks {
			time.Sleep(time.Millisecond)
		}
	}
}

// closeDeferred marks the file closing. If goroutines are waiting on the fd it
// reports true, and the last of them closes the fd so it is not reused early.
func (f *file) closeDeferred() bool {
	s := &f.nb
	s.mu.Lock()
	defer s.mu.Unlock()
	s.closing = true
	return s.waiters > 0
}

const (
	pollIn  = 0x1
	pollOut = 0x4
)

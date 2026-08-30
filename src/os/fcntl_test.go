//go:build (darwin || linux) && !baremetal && !tinygo.wasm && !nintendoswitch

package os_test

import (
	. "os"
	"syscall"
	"testing"
	"time"
)

// fcntl takes its third parameter through a variadic list, so the call needs
// the C wrapper in src/runtime/os_darwin.c on darwin.

// F_SETFL must reach fcntl with the value that the caller gave it. A read on
// an empty pipe returns EAGAIN when the descriptor is non-blocking, and blocks
// when the flag did not arrive.
func TestFcntlSetNonblock(t *testing.T) {
	var fds [2]int
	if err := syscall.Pipe(fds[:]); err != nil {
		t.Fatalf("Pipe failed: %v", err)
	}
	defer syscall.Close(fds[0])
	defer syscall.Close(fds[1])

	if err := syscall.SetNonblock(fds[0], true); err != nil {
		t.Fatalf("SetNonblock failed: %v", err)
	}

	type result struct {
		n   int
		err error
	}
	done := make(chan result, 1)
	go func() {
		buf := make([]byte, 1)
		n, err := syscall.Read(fds[0], buf)
		done <- result{n, err}
	}()

	select {
	case r := <-done:
		if r.err != syscall.EAGAIN {
			t.Errorf("wanted EAGAIN from a read on an empty non-blocking pipe, got %d, %v", r.n, r.err)
		}
	case <-time.After(10 * time.Second):
		t.Fatal("the read blocked, so the descriptor is still blocking")
	}
}

// The pointer commands go through the same wrapper as the int commands. A
// F_GETLK on a file that nobody locked reports F_UNLCK.
func TestFcntlGetLock(t *testing.T) {
	f, err := CreateTemp(t.TempDir(), "fcntl")
	if err != nil {
		t.Fatalf("CreateTemp failed: %v", err)
	}
	defer f.Close()

	lk := syscall.Flock_t{
		Type:   syscall.F_RDLCK,
		Whence: 0,
		Start:  0,
		Len:    0,
	}
	if err := syscall.FcntlFlock(f.Fd(), syscall.F_GETLK, &lk); err != nil {
		t.Fatalf("FcntlFlock(F_GETLK) failed: %v", err)
	}
	if lk.Type != syscall.F_UNLCK {
		t.Errorf("wanted F_UNLCK on an unlocked file, got %d", lk.Type)
	}
}

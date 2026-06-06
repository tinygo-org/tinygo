//go:build wasip1

// Internal-test helpers exposed via go:linkname so user code can drive
// the deadline-aware Read/Write loop without becoming a stdlib package.
// Not part of the public API; the names are intentionally awkward to
// signal "for tests only".

package poll

import (
	"syscall"
	"time"
)

// pollTestReadWithDeadline opens a pollable FD wrapper for sysfd, sets
// a read deadline d into the future, calls Read once, and returns
// (n, err). Caller is responsible for closing sysfd.
//
//go:linkname pollTestReadWithDeadline
func pollTestReadWithDeadline(sysfd int, d time.Duration, p []byte) (int, error) {
	fd := &FD{Sysfd: sysfd, IsStream: true}
	// Best-effort init; ignore error so a caller using a not-fcntl-able FD
	// (stdin under wazero, etc.) still gets to test the deadline path on
	// whatever park behaviour the runtime gives.
	_ = fd.Init("test", true)
	if err := fd.SetReadDeadline(time.Now().Add(d)); err != nil {
		return 0, err
	}
	return fd.Read(p)
}

// pollTestSetNonblock toggles O_NONBLOCK on a raw sysfd. Useful in
// tests when the caller wants to ensure the FD is in nonblocking mode
// before calling pollTestReadWithDeadline (Init is best-effort and may
// silently skip).
//
//go:linkname pollTestSetNonblock
func pollTestSetNonblock(sysfd int) error {
	flags, err := syscall.Fcntl(sysfd, syscall.F_GETFL, 0)
	if err != nil {
		return err
	}
	_, err = syscall.Fcntl(sysfd, syscall.F_SETFL, flags|syscall.O_NONBLOCK)
	return err
}

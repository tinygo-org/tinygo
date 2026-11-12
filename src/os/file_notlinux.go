//go:build !linux || baremetal || tinygo.wasm || nintendoswitch

package os

import (
	"syscall"
	"time"
)

//TODO: re-implement the ErrNoDeadline error in the correct code path

func (f *File) read(b []byte) (int, error) {
	return f.handle.Read(b)
}

func (f *File) write(b []byte) (int, error) {
	return f.handle.Write(b)
}

func (f *File) syscallConn() (syscall.RawConn, error) {
	return nil, ErrNotImplemented
}

// setDeadline sets the read and write deadline.
func (f *File) setDeadline(t time.Time) error {
	if t.IsZero() {
		return nil
	}
	return ErrNotImplemented
}

// setReadDeadline sets the read deadline, not yet implemented
// A zero value for t means Read will not time out.
func (f *File) setReadDeadline(t time.Time) error {
	if t.IsZero() {
		return nil
	}
	return ErrNotImplemented
}

// setWriteDeadline sets the write deadline, not yet implemented
// A zero value for t means Read will not time out.
func (f *File) setWriteDeadline(t time.Time) error {
	if t.IsZero() {
		return nil
	}
	return ErrNotImplemented
}

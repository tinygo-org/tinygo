package os

import (
	"time"
)

//TODO: re-implement the ErrNoDeadline error in the correct code path

// Chtimes is a stub, not yet implemented
func Chtimes(name string, atime time.Time, mtime time.Time) error {
	return ErrNotImplemented
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

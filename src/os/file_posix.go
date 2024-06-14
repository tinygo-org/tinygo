package os

import (
	"time"
)

// Chtimes is a stub, not yet implemented
func Chtimes(name string, atime time.Time, mtime time.Time) error {
	return ErrNotImplemented
}

// setReadDeadline sets the read deadline, not yet implemented
func (f *File) setReadDeadline(_ time.Time) error {
	return ErrNotImplemented
}

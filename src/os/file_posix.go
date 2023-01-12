package os

import (
	"time"
)

// Chtimes is a stub, not yet implemented
func Chtimes(name string, atime time.Time, mtime time.Time) error {
	return ErrNotImplemented
}

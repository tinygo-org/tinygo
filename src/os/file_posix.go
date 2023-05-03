//go:build !windows

package os

import (
	"syscall"
	"time"
)

// Chtimes is a stub, not yet implemented
func Chtimes(name string, atime time.Time, mtime time.Time) error {
	return ErrNotImplemented
}

// Truncate changes the size of the named file.
// If the file is a symbolic link, it changes the size of the link's target.
// If there is an error, it will be of type *PathError.
func Truncate(path string, size int64) error {
	err := syscall.Truncate(path, size)
	if err != nil {
		err = &PathError{Op: "truncate", Path: path, Err: err}
	}
	return err
}

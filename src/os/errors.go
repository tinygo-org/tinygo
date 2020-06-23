package os

import (
	"errors"
)

var (
	ErrInvalid    = errors.New("invalid argument")
	ErrPermission = errors.New("permission denied")
	ErrClosed     = errors.New("file already closed")

	// Portable analogs of some common system call errors.
	// Note that these are exported for use in the Filesystem interface.
	ErrUnsupported    = errors.New("operation not supported")
	ErrNotImplemented = errors.New("operation not implemented")
	ErrNotExist       = errors.New("file not found")
	ErrExist          = errors.New("file exists")
)

func IsPermission(err error) bool {
	return err == ErrPermission
}

func NewSyscallError(syscall string, err error) error {
	if err == nil {
		return nil
	}
	return &SyscallError{syscall, err}
}

// SyscallError records an error from a specific system call.
type SyscallError struct {
	Syscall string
	Err     error
}

func (e *SyscallError) Error() string { return e.Syscall + ": " + e.Err.Error() }

func (e *SyscallError) Unwrap() error { return e.Err }

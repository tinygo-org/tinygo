package os

import (
	"errors"
	"syscall"
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

// The following code is copied from the official implementation.
// src/internal/poll/fd.go

// ErrDeadlineExceeded is returned for an expired deadline.
// This is exported by the os package as os.ErrDeadlineExceeded.
var ErrDeadlineExceeded error = &DeadlineExceededError{}

// DeadlineExceededError is returned for an expired deadline.
type DeadlineExceededError struct{}

// Implement the net.Error interface.
// The string is "i/o timeout" because that is what was returned
// by earlier Go versions. Changing it may break programs that
// match on error strings.
func (e *DeadlineExceededError) Error() string   { return "i/o timeout" }
func (e *DeadlineExceededError) Timeout() bool   { return true }
func (e *DeadlineExceededError) Temporary() bool { return true }

// The following code is copied from the official implementation.
// https://github.com/golang/go/blob/4ce6a8e89668b87dce67e2f55802903d6eb9110a/src/os/error.go#L65-L104

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

func IsExist(err error) bool {
	return underlyingErrorIs(err, ErrExist)
}

func IsNotExist(err error) bool {
	return underlyingErrorIs(err, ErrNotExist)
}

func IsPermission(err error) bool {
	return underlyingErrorIs(err, ErrPermission)
}

func underlyingErrorIs(err, target error) bool {
	// Note that this function is not errors.Is:
	// underlyingError only unwraps the specific error-wrapping types
	// that it historically did, not all errors implementing Unwrap().
	err = underlyingError(err)
	if err == target {
		return true
	}
	// To preserve prior behavior, only examine syscall errors.
	e, ok := err.(syscall.Errno)
	return ok && e.Is(target)
}

// underlyingError returns the underlying error for known os error types.
func underlyingError(err error) error {
	switch err := err.(type) {
	case *PathError:
		return err.Err
	case *SyscallError:
		return err.Err
	}
	return err
}

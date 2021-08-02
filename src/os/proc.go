// Package os implements a subset of the Go "os" package. See
// https://godoc.org/os for details.
//
// Note that the current implementation is blocking. This limitation should be
// removed in a future version.
package os

import (
	"syscall"
)

// Args hold the command-line arguments, starting with the program name.
var Args []string

func init() {
	Args = runtime_args()
}

func runtime_args() []string // in package runtime

// Exit causes the current program to exit with the given status code.
// Conventionally, code zero indicates success, non-zero an error.
// The program terminates immediately; deferred functions are not run.
func Exit(code int) {
	syscall.Exit(code)
}

// Getuid returns the numeric user id of the caller.
//
// On non-POSIX systems, it returns -1.
func Getuid() int {
	return syscall.Getuid()
}

// Geteuid returns the numeric effective user id of the caller.
//
// On non-POSIX systems, it returns -1.
func Geteuid() int {
	return syscall.Geteuid()
}

// Getgid returns the numeric group id of the caller.
//
// On non-POSIX systems, it returns -1.
func Getgid() int {
	return syscall.Getgid()
}

// Getegid returns the numeric effective group id of the caller.
//
// On non-POSIX systems, it returns -1.
func Getegid() int {
	return syscall.Getegid()
}

package os

import "syscall"

// Getpagesize returns the underlying system's memory page size.
func Getpagesize() int { return syscall.Getpagesize() }

//go:build tinygo.wasi || js

package syscall

// wasi-libc's errno is thread-local, and cgo code that sets it broke a plain
// extern errno, so it is reached through __errno_location instead.
//
//export __errno_location
func libc_errno_location() *Errno

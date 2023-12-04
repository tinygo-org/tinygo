//go:build wasip1

package syscall

// Use a go:extern definition to access the errno from wasi-libc
//
//go:extern errno
var libcErrno Errno

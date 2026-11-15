//go:build tinygo.wasi || js

package syscall

// Use a go:extern definition to access the errno from wasi-libc
//
//go:extern errno
var libcErrno Errno

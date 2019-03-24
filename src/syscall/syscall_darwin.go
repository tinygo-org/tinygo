package syscall

// This file defines errno and constants to match the darwin libsystem ABI.
// Values have been determined experimentally by compiling some C code on macOS
// with Clang and looking at the resulting LLVM IR.

// This function returns the error location in the darwin ABI.
// Discovered by compiling the following code using Clang:
//
//     #include <errno.h>
//     int getErrno() {
//         return errno;
//     }
//
//go:export __error
func libc___error() *int32

// getErrno returns the current C errno. It may not have been caused by the last
// call, so it should only be relied upon when the last call indicates an error
// (for example, by returning -1).
func getErrno() Errno {
	errptr := libc___error()
	return Errno(uintptr(*errptr))
}

const (
	ENOENT      Errno = 2
	EINTR       Errno = 4
	EMFILE      Errno = 24
	EAGAIN      Errno = 35
	ETIMEDOUT   Errno = 60
	ENOSYS      Errno = 78
	EWOULDBLOCK Errno = EAGAIN
)

type Signal int

const (
	SIGCHLD Signal = 20
	SIGINT  Signal = 2
	SIGKILL Signal = 9
	SIGTRAP Signal = 5
	SIGQUIT Signal = 3
	SIGTERM Signal = 15
)

const (
	O_RDONLY = 0
	O_WRONLY = 1
	O_RDWR   = 2
)

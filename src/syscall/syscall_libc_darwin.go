// +build darwin

package syscall

// This file defines errno and constants to match the darwin libsystem ABI.
// Values have been copied from src/syscall/zerrors_darwin_amd64.go.

// This function returns the error location in the darwin ABI.
// Discovered by compiling the following code using Clang:
//
//     #include <errno.h>
//     int getErrno() {
//         return errno;
//     }
//
//export __error
func libc___error() *int32

// getErrno returns the current C errno. It may not have been caused by the last
// call, so it should only be relied upon when the last call indicates an error
// (for example, by returning -1).
func getErrno() Errno {
	errptr := libc___error()
	return Errno(uintptr(*errptr))
}

func (e Errno) Is(target error) bool {
	switch target.Error() {
	case "permission denied":
		return e == EACCES || e == EPERM
	case "file already exists":
		return e == EEXIST
	case "file does not exist":
		return e == ENOENT
	}
	return false
}

const (
	EPERM       Errno = 0x1
	ENOENT      Errno = 0x2
	EACCES      Errno = 0xd
	EEXIST      Errno = 0x11
	EINTR       Errno = 0x4
	ENOTDIR     Errno = 0x14
	EINVAL      Errno = 0x16
	EMFILE      Errno = 0x18
	EAGAIN      Errno = 0x23
	ETIMEDOUT   Errno = 0x3c
	ENOSYS      Errno = 0x4e
	EWOULDBLOCK Errno = EAGAIN
)

type Signal int

const (
	SIGCHLD Signal = 0x14
	SIGINT  Signal = 0x2
	SIGKILL Signal = 0x9
	SIGTRAP Signal = 0x5
	SIGQUIT Signal = 0x3
	SIGTERM Signal = 0xf
)

const (
	Stdin  = 0
	Stdout = 1
	Stderr = 2
)

const (
	O_RDONLY = 0x0
	O_WRONLY = 0x1
	O_RDWR   = 0x2
	O_APPEND = 0x8
	O_SYNC   = 0x80
	O_CREAT  = 0x200
	O_TRUNC  = 0x400
	O_EXCL   = 0x800
)

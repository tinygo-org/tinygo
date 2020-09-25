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

// Source: https://opensource.apple.com/source/xnu/xnu-7195.81.3/bsd/sys/errno.h.auto.html
const (
	EPERM       Errno = 1
	ENOENT      Errno = 2
	EACCES      Errno = 13
	EEXIST      Errno = 17
	EINTR       Errno = 4
	ENOTDIR     Errno = 20
	EINVAL      Errno = 22
	EMFILE      Errno = 24
	EPIPE       Errno = 32
	EAGAIN      Errno = 35
	ETIMEDOUT   Errno = 60
	ENOSYS      Errno = 78
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

	O_CLOEXEC = 0x01000000
)

// Source: https://opensource.apple.com/source/xnu/xnu-7195.81.3/bsd/sys/mman.h.auto.html
const (
	PROT_NONE  = 0x00 // no permissions
	PROT_READ  = 0x01 // pages can be read
	PROT_WRITE = 0x02 // pages can be written
	PROT_EXEC  = 0x04 // pages can be executed

	MAP_SHARED  = 0x0001 // share changes
	MAP_PRIVATE = 0x0002 // changes are private

	MAP_FILE      = 0x0000 // map from file (default)
	MAP_ANON      = 0x1000 // allocated from memory, swap space
	MAP_ANONYMOUS = MAP_ANON
)

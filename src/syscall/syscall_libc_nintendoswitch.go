//go:build nintendoswitch

package syscall

import (
	"internal/itoa"
)

// A Signal is a number describing a process signal.
// It implements the os.Signal interface.
type Signal int

const (
	_ Signal = iota
	SIGCHLD
	SIGINT
	SIGKILL
	SIGTRAP
	SIGQUIT
	SIGTERM
)

func (s Signal) Signal() {}

func (s Signal) String() string {
	if 0 <= s && int(s) < len(signals) {
		str := signals[s]
		if str != "" {
			return str
		}
	}
	return "signal " + itoa.Itoa(int(s))
}

var signals = [...]string{}

// File system

const (
	Stdin  = 0
	Stdout = 1
	Stderr = 2
)

const (
	O_RDONLY = 0
	O_WRONLY = 1
	O_RDWR   = 2

	O_CREAT  = 0100
	O_TRUNC  = 01000
	O_APPEND = 02000
	O_EXCL   = 0200
	O_SYNC   = 010000

	O_CLOEXEC = 0
)

//go:extern errno
var libcErrno uintptr

func getErrno() error {
	return Errno(libcErrno)
}

func Pipe2(p []int, flags int) (err error) {
	return ENOSYS // TODO
}

func Getpagesize() int {
	return 4096 // TODO
}

type RawSockaddrInet4 struct {
	// stub
}

type RawSockaddrInet6 struct {
	// stub
}

func Chmod(path string, mode uint32) (err error) {
	data := cstring(path)
	fail := int(libc_chmod(&data[0], mode))
	if fail < 0 {
		err = getErrno()
	}
	return
}

// int open(const char *pathname, int flags, mode_t mode);
//
//export open
func libc_open(pathname *byte, flags int32, mode uint32) int32

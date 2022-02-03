//go:build nintendoswitch
// +build nintendoswitch

package syscall

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

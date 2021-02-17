// +build wasi

package syscall

// https://github.com/WebAssembly/wasi-libc/blob/main/expected/wasm32-wasi/predefined-macros.txt

type Signal int

const (
	SIGCHLD = 16
	SIGINT  = 2
	SIGKILL = 9
	SIGTRAP = 5
	SIGQUIT = 3
	SIGTERM = 15
)

const (
	Stdin  = 0
	Stdout = 1
	Stderr = 2
)

const (
	__WASI_OFLAGS_CREAT   = 1
	__WASI_FDFLAGS_APPEND = 1
	__WASI_OFLAGS_EXCL    = 4
	__WASI_OFLAGS_TRUNC   = 8
	__WASI_FDFLAGS_SYNC   = 16

	O_RDONLY = 0x04000000
	O_WRONLY = 0x10000000
	O_RDWR   = O_RDONLY | O_WRONLY

	O_CREAT  = __WASI_OFLAGS_CREAT << 12
	O_CREATE = O_CREAT
	O_TRUNC  = __WASI_OFLAGS_TRUNC << 12
	O_APPEND = __WASI_FDFLAGS_APPEND
	O_EXCL   = __WASI_OFLAGS_EXCL << 12
	O_SYNC   = __WASI_FDFLAGS_SYNC

	O_CLOEXEC = 0
)

//go:extern errno
var libcErrno uintptr

func getErrno() error {
	return Errno(libcErrno)
}

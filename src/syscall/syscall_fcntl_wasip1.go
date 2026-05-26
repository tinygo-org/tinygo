//go:build wasip1

package syscall

import "unsafe"

// __wasi_fdstat_t mirrors the wasip1 fdstat record. Per the spec
// (https://github.com/WebAssembly/WASI/blob/main/legacy/preview1/docs.md#-fdstat-record):
//
//	size: 24, align: 8
//	fs_filetype:           u8  at offset 0
//	fs_flags:              u16 at offset 2
//	fs_rights_base:        u64 at offset 8
//	fs_rights_inheriting:  u64 at offset 16
type __wasi_fdstat_t struct {
	fsFiletype         uint8
	_                  uint8
	fsFlags            uint16
	_                  [4]byte
	fsRightsBase       uint64
	fsRightsInheriting uint64
}

var _ [0]byte = [24 - unsafe.Sizeof(__wasi_fdstat_t{})]byte{}

//go:wasmimport wasi_snapshot_preview1 fd_fdstat_get
func fd_fdstat_get(fd int32, out *__wasi_fdstat_t) uint16

//go:wasmimport wasi_snapshot_preview1 fd_fdstat_set_flags
func fd_fdstat_set_flags(fd int32, flags uint16) uint16

// Fcntl is a minimal subset of POSIX fcntl backed by wasip1's fd_fdstat
// primitives. Only F_GETFL and F_SETFL are supported on wasip1 (these are
// the only commands TinyGo's runtime needs for setting O_NONBLOCK). The
// libc fcntl path can't be used because wasi-libc's fcntl is variadic and
// the Go wasmimport binding has no way to express that.
func Fcntl(fd int, cmd int, arg int) (val int, err error) {
	switch cmd {
	case F_GETFL:
		var st __wasi_fdstat_t
		if errno := fd_fdstat_get(int32(fd), &st); errno != 0 {
			err = Errno(errno)
			return
		}
		return int(st.fsFlags), nil
	case F_SETFL:
		if errno := fd_fdstat_set_flags(int32(fd), uint16(arg)); errno != 0 {
			err = Errno(errno)
			return
		}
		return 0, nil
	default:
		err = ENOSYS
		return
	}
}

// Filetype is the wasi filetype tag returned by fd_fdstat_get for any
// open file descriptor. Used by upstream net/file_wasip1.go to decide
// whether a pre-opened FD should be wrapped as net.Listener (stream
// socket) or net.Conn (stream / dgram socket).
type Filetype = uint8

const (
	FILETYPE_UNKNOWN          Filetype = 0
	FILETYPE_BLOCK_DEVICE     Filetype = 1
	FILETYPE_CHARACTER_DEVICE Filetype = 2
	FILETYPE_DIRECTORY        Filetype = 3
	FILETYPE_REGULAR_FILE     Filetype = 4
	FILETYPE_SOCKET_DGRAM     Filetype = 5
	FILETYPE_SOCKET_STREAM    Filetype = 6
	FILETYPE_SYMBOLIC_LINK    Filetype = 7
)

// fd_fdstat_get_type returns the wasi filetype of fd. Used by upstream
// Go's net/file_wasip1.go via //go:linkname syscall.fd_fdstat_get_type
// to detect socket FDs handed in by the host runtime.
//
//go:linkname fd_fdstat_get_type
func fd_fdstat_get_type(fd int) (Filetype, error) {
	var st __wasi_fdstat_t
	if errno := fd_fdstat_get(int32(fd), &st); errno != 0 {
		return 0, Errno(errno)
	}
	return st.fsFiletype, nil
}

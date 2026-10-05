//go:build wasip1

package syscall

import "unsafe"

func firstPreopenDirectory() string {
	for fd := int32(3); ; fd++ {
		var stat wasiPrestat
		errno := Errno(fd_prestat_get(fd, &stat))
		if errno == EBADF {
			return ""
		}
		if errno == ENOTDIR {
			continue
		}
		if errno != 0 {
			panic("fd_prestat_get: " + errno.Error())
		}
		if stat.typ != 0 {
			continue
		}
		name := make([]byte, stat.nameLen)
		if errno := Errno(fd_prestat_dir_name(fd, unsafe.SliceData(name), stat.nameLen)); errno != 0 {
			panic("fd_prestat_dir_name: " + errno.Error())
		}
		return string(name)
	}
}

type wasiPrestat struct {
	typ     uint8
	_       [3]byte
	nameLen uint32
}

//go:wasmimport wasi_snapshot_preview1 fd_prestat_get
//go:noescape
func fd_prestat_get(fd int32, stat *wasiPrestat) uint32

//go:wasmimport wasi_snapshot_preview1 fd_prestat_dir_name
//go:noescape
func fd_prestat_dir_name(fd int32, name *byte, size uint32) uint32

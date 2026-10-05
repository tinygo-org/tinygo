//go:build wasip1

package syscall

import (
	"path"
	_ "runtime" // Run libc constructors in runtime_wasip1.go before reading PWD.
	"unsafe"
)

// Defined in wasi-libc's libc-bottom-half/sources/getcwd.c.
//
//go:extern __wasilibc_cwd
var libcCwd *byte

var initialWorkingDir string
var initialWorkingDirBuffer []byte

func init() {
	if pwd, _ := Getenv("PWD"); pwd != "" {
		initialWorkingDir = path.Join("/", pwd)
		if pwd[len(pwd)-1] == '/' && initialWorkingDir[len(initialWorkingDir)-1] != '/' {
			initialWorkingDir += "/"
		}
	} else {
		initialWorkingDir = firstPreopenDirectory()
	}
	cwd := initialWorkingDir
	if cwd == "" {
		cwd = "/"
	}
	initialWorkingDirBuffer = cstring(cwd)
	libcCwd = unsafe.SliceData(initialWorkingDirBuffer)
}

func Getwd() (string, error) {
	if libcCwd == unsafe.SliceData(initialWorkingDirBuffer) {
		return initialWorkingDir, nil
	}
	return string(unsafe.Slice(libcCwd, libc_strlen(unsafe.Pointer(libcCwd)))), nil
}

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

//go:build wasip1 || wasip2

package syscall

import (
	"path"
	_ "runtime" // Run libc constructors before reading PWD.
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
	// libc does not free its initial cwd on the first chdir.
	// See lib/wasi-libc/libc-bottom-half/sources/chdir.c (chdir).
	initialWorkingDirBuffer = cstring(cwd)
	libcCwd = unsafe.SliceData(initialWorkingDirBuffer)
}

func Getwd() (string, error) {
	if libcCwd == unsafe.SliceData(initialWorkingDirBuffer) {
		return initialWorkingDir, nil
	}
	return string(unsafe.Slice(libcCwd, libc_strlen(unsafe.Pointer(libcCwd)))), nil
}

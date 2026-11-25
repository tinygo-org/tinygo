//go:build linux && !baremetal && !tinygo.wasm && (386 || arm)

package os

import (
	"syscall"
	"unsafe"
)

//export syscall.seek
func seek(fd int, offset int64, whence int) (newoffset int64, err syscall.Errno) {
	_, _, err = syscall.Syscall6(syscall.SYS__LLSEEK, uintptr(fd),
		uintptr(uint64(offset)>>32), uintptr(uint32(offset)),
		uintptr(unsafe.Pointer(&newoffset)), uintptr(whence), 0)
	if err != 0 {
		return 0, err
	}
	return newoffset, 0
}

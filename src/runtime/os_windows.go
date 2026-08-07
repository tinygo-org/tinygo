package runtime

import "unsafe"

const GOOS = "windows"

const zeroSizeAllocPtr uintptr = 16 // part of the first protected page

type systeminfo struct {
	anon0                       [4]byte
	dwpagesize                  uint32
	lpminimumapplicationaddress *byte
	lpmaximumapplicationaddress *byte
	dwactiveprocessormask       uintptr
	dwnumberofprocessors        uint32
	dwprocessortype             uint32
	dwallocationgranularity     uint32
	wprocessorlevel             uint16
	wprocessorrevision          uint16
}

//export GetSystemInfo
func _GetSystemInfo(lpSystemInfo unsafe.Pointer)

//go:linkname syscall_Getpagesize syscall.Getpagesize
func syscall_Getpagesize() int {
	var info systeminfo
	_GetSystemInfo(unsafe.Pointer(&info))
	return int(info.dwpagesize)
}

//export _errno
func libc_errno_location() *int32

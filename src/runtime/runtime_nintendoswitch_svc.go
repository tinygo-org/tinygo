// +build nintendoswitch

package runtime

import (
	"unsafe"
)

// Result nxOutputString(const char *str, u64 size)
//export nxOutputString
func nxOutputString(str *uint8, size uint64) uint64

func NxOutputString(str string) uint64 {
	strData := (*_string)(unsafe.Pointer(&str))
	return nxOutputString((*uint8)(unsafe.Pointer(strData.ptr)), uint64(strData.length))
}

//export malloc
func extalloc(size uintptr) unsafe.Pointer

//export free
func extfree(ptr unsafe.Pointer)

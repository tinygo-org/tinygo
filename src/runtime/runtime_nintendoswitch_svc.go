// +build nintendoswitch

package runtime

import (
	"unsafe"
)

// Result svcOutputDebugString(const char *str, u64 size)
//go:export svcOutputDebugString
func svcOutputDebugString(str *byte, size uint64) uint64

//go:export malloc
func extalloc(size uintptr) unsafe.Pointer

//export free
func extfree(ptr unsafe.Pointer)

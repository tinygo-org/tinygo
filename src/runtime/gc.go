// +build !linux

package runtime

import (
	"unsafe"
)

var (
	_extern__heap_start unsafe.Pointer // defined by the linker
	heapptr             = uintptr(unsafe.Pointer(&_extern__heap_start))
)

func alloc(size uintptr) unsafe.Pointer {
	// TODO: this can be optimized by not casting between pointers and ints so
	// much. And by using platform-native data types (e.g. *uint8 for 8-bit
	// systems).
	size = align(size)
	addr := heapptr
	heapptr += size
	for i := uintptr(0); i < uintptr(size); i += 4 {
		ptr := (*uint32)(unsafe.Pointer(addr + i))
		*ptr = 0
	}
	return unsafe.Pointer(addr)
}

func free(ptr unsafe.Pointer) {
	// TODO: use a GC
}

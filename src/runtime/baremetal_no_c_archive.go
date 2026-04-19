//go:build baremetal && !buildmode.c_archive

package runtime

import "unsafe"

//go:extern _heap_start
var heapStartSymbol [0]byte

//go:extern _heap_end
var heapEndSymbol [0]byte

//go:extern _globals_start
var globalsStartSymbol [0]byte

//go:extern _globals_end
var globalsEndSymbol [0]byte

//go:extern _stack_top
var stackTopSymbol [0]byte

var (
	heapStart    = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd      = uintptr(unsafe.Pointer(&heapEndSymbol))
	globalsStart = uintptr(unsafe.Pointer(&globalsStartSymbol))
	globalsEnd   = uintptr(unsafe.Pointer(&globalsEndSymbol))
	stackTop     = uintptr(unsafe.Pointer(&stackTopSymbol))
)

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	// Note: this zeroes the returned buffer which is not necessary.
	// The same goes for bytealg.MakeNoZero.
	return alloc(size, nil)
}

//export calloc
func libc_calloc(nmemb, size uintptr) unsafe.Pointer {
	// No difference between calloc and malloc.
	return libc_malloc(nmemb * size)
}

//export free
func libc_free(ptr unsafe.Pointer) {
	free(ptr)
}

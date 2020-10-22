// +build baremetal

package runtime

import (
	"unsafe"
)

//go:extern _heap_start
var heapStartSymbol [0]byte

//go:extern _heap_end
var heapEndSymbol [0]byte

//go:extern _globals_start
var globalsStartSymbol [0]byte

//go:extern _globals_end
var globalsEndSymbol [0]byte

//go:extern _stack_size
var stackSizeSymbol [0]byte

//go:extern _stack_top
var stackTopSymbol [0]byte

var (
	heapStart    = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd      = uintptr(unsafe.Pointer(&heapEndSymbol))
	globalsStart = uintptr(unsafe.Pointer(&globalsStartSymbol))
	globalsEnd   = uintptr(unsafe.Pointer(&globalsEndSymbol))
	stackSize    = uintptr(unsafe.Pointer(&stackSizeSymbol))
	stackTop     = uintptr(unsafe.Pointer(&stackTopSymbol))
)

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	return alloc(size)
}

//export free
func libc_free(ptr unsafe.Pointer) {
	free(ptr)
}

//go:linkname syscall_Exit syscall.Exit
func syscall_Exit(code int) {
	abort()
}

const baremetal = true

// +build nintendoswitch

// +build gc.conservative gc.leaking

package runtime

import "unsafe"

const heapSize = 16 * 1024 * 1024

//go:extern _stack_top
var stackTopSymbol [0]byte

var (
	heapStart = uintptr(0)
	heapEnd   = uintptr(0)
	stackTop  = uintptr(unsafe.Pointer(&stackTopSymbol))
)

func preinit() {
	heapStart = uintptr(extalloc(heapSize))

	if heapStart == 0 {
		panic("Cannot allocate heap")
	}

	heapEnd = heapStart + heapSize
}

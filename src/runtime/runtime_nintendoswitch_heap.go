// +build nintendoswitch

// +build gc.conservative gc.leaking

package runtime

import "unsafe"

const heapSize = 0x2000000 * 16 // Default by libnx

//go:extern _stack_top
var stackTopSymbol [0]byte

var (
	heapStart = uintptr(0)
	heapEnd   = uintptr(0)
	stackTop  = uintptr(unsafe.Pointer(&stackTopSymbol))
)

//export setHeapSize
func setHeapSize(addr *uintptr, length uint64) uint64

func preinit() {
	setHeapSize(&heapStart, heapSize)

	if heapStart == 0 {
		panic("failed to allocate heap")
	}

	heapEnd = heapStart + heapSize
}

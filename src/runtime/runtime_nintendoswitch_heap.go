// +build nintendoswitch

// +build gc.conservative gc.leaking

package runtime

const heapSize = 0x2000000 * 16 // Default by libnx

var (
	heapStart = uintptr(0)
	heapEnd   = uintptr(0)
)

//export setHeapSize
func setHeapSize(addr *uintptr, length uint64) uint64

func preinit() {
	setHeapSize(&heapStart, heapSize)

	if heapStart == 0 {
		runtimePanic("failed to allocate heap")
	}

	heapEnd = heapStart + heapSize
}

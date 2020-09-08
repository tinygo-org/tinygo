// +build nintendoswitch

// +build gc.conservative gc.leaking

package runtime

const heapSize = 0x2000000 * 16 // Default by libnx

var (
	heapStart = uintptr(0)
	heapEnd   = uintptr(0)
	failedToAllocateHeapString = []byte("failed to allocate heap")
)

//export setHeapSize
func setHeapSize(addr *uintptr, length uint64) uint64

func preinit() {
	setHeapSize(&heapStart, heapSize)

	if heapStart == 0 {
		// Panic should work, but if we call panic here, it triggers a compiler bug that breaks
		// all print calls in the whole application. So far we couldn't figure out why, that's
		// why we're using a byte array and nxOutputString directly
		nxOutputString(&failedToAllocateHeapString[0], uint64(len(failedToAllocateHeapString)))
		exit(1)
		//panic("failed to allocate heap")
	}

	heapEnd = heapStart + heapSize
}

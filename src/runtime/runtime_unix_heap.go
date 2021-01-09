// +build darwin linux,!baremetal,!wasi freebsd,!baremetal
// +build !nintendoswitch

// +build gc.conservative gc.leaking

package runtime

const heapSize = 1 * 1024 * 1024 // 1MB to start

var heapStart, heapEnd uintptr

func preinit() {
	heapStart = uintptr(malloc(heapSize))
	heapEnd = heapStart + heapSize
}

// growHeap tries to grow the heap size. It returns true if it succeeds, false
// otherwise.
func growHeap() bool {
	// At the moment, this is not possible. However it shouldn't be too
	// difficult (at least on Linux) to allocate a large amount of virtual
	// memory at startup that is then slowly used.
	return false
}

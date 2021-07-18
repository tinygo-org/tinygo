// +build darwin linux,!baremetal,!wasi freebsd,!baremetal
// +build !nintendoswitch

// +build gc.conservative gc.leaking gc.list

package runtime

var heapSize uintptr = 128 * 1024          // small amount to start
const heapMaxSize = 1 * 1024 * 1024 * 1024 // 1GB for the entire heap

var heapStart, heapEnd uintptr

func preinit() {
	// Allocate a large chunk of virtual memory. Because it is virtual, it won't
	// really be allocated in RAM. Memory will only be allocated when it is
	// first touched.
	addr := mmap(nil, heapMaxSize, flag_PROT_READ|flag_PROT_WRITE, flag_MAP_PRIVATE|flag_MAP_ANONYMOUS, -1, 0)
	heapStart = uintptr(addr)
	heapEnd = heapStart + heapSize
}

// growHeap tries to grow the heap size. It returns true if it succeeds, false
// otherwise.
func growHeap() bool {
	if heapSize == heapMaxSize {
		// Already at the max. If we run out of memory, we should consider
		// increasing heapMaxSize on 64-bit systems.
		return false
	}
	// Grow the heap size used by the program.
	heapSize = (heapSize * 4 / 3) &^ 4095 // grow by around 33%
	if heapSize > heapMaxSize {
		heapSize = heapMaxSize
	}
	setHeapEnd(heapStart + heapSize)
	return true
}

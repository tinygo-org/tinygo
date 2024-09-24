//go:build gc.leaking

package runtime

// This GC implementation is the simplest useful memory allocator possible: it
// only allocates memory and never frees it. For some constrained systems, it
// may be the only memory allocator possible.

import (
	"unsafe"
)

const needsStaticHeap = true

// Ever-incrementing pointer: no memory is freed.
var heapptr = heapStart

// Total amount allocated for runtime.MemStats
var gcTotalAlloc uint64

// Total number of calls to alloc()
var gcMallocs uint64

// Total number of objected freed; for leaking collector this stays 0
const gcFrees = 0

// Inlining alloc() speeds things up slightly but bloats the executable by 50%,
// see https://github.com/tinygo-org/tinygo/issues/2674.  So don't.
//
//go:noinline
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer {
	// TODO: this can be optimized by not casting between pointers and ints so
	// much. And by using platform-native data types (e.g. *uint8 for 8-bit
	// systems).
	size = align(size)
	addr := heapptr
	gcTotalAlloc += uint64(size)
	gcMallocs++
	heapptr += size
	for heapptr >= heapEnd {
		// Try to increase the heap and check again.
		if growHeap() {
			continue
		}
		// Failed to make the heap bigger, so we must really be out of memory.
		runtimePanic("out of memory")
	}
	pointer := unsafe.Pointer(addr)
	zero_new_alloc(pointer, size)
	return pointer
}

func realloc(ptr unsafe.Pointer, size uintptr) unsafe.Pointer {
	newAlloc := alloc(size, nil)
	if ptr == nil {
		return newAlloc
	}
	// according to POSIX everything beyond the previous pointer's
	// size will have indeterminate values so we can just copy garbage
	memcpy(newAlloc, ptr, size)

	return newAlloc
}

func free(ptr unsafe.Pointer) {
	// Memory is never freed.
}

// ReadMemStats populates m with memory statistics.
//
// The returned memory statistics are up to date as of the
// call to ReadMemStats. This would not do GC implicitly for you.
func ReadMemStats(m *MemStats) {
	m.HeapIdle = 0
	m.HeapInuse = gcTotalAlloc
	m.HeapReleased = 0 // always 0, we don't currently release memory back to the OS.

	m.HeapSys = m.HeapInuse + m.HeapIdle
	m.GCSys = 0
	m.TotalAlloc = gcTotalAlloc
	m.Mallocs = gcMallocs
	m.Frees = gcFrees
	m.Sys = uint64(heapEnd - heapStart)
}

func GC() {
	// No-op.
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// No-op.
}

func initHeap() {
	// preinit() may have moved heapStart; reset heapptr
	heapptr = heapStart
}

// setHeapEnd sets a new (larger) heapEnd pointer.
func setHeapEnd(newHeapEnd uintptr) {
	// This "heap" is so simple that simply assigning a new value is good
	// enough.
	heapEnd = newHeapEnd
}

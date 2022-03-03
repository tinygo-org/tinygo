//go:build gc.leaking
// +build gc.leaking

package runtime

// This GC implementation is the simplest useful memory allocator possible: it
// only allocates memory and never frees it. For some constrained systems, it
// may be the only memory allocator possible.

import (
	"unsafe"
)

// Ever-incrementing pointer: no memory is freed.
var heapptr = heapStart

func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer {
	// TODO: this can be optimized by not casting between pointers and ints so
	// much. And by using platform-native data types (e.g. *uint8 for 8-bit
	// systems).
	size = align(size)
	addr := heapptr
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
	memzero(pointer, size)
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

func GC() {
	// No-op.
}

func KeepAlive(x interface{}) {
	// Unimplemented. Only required with SetFinalizer().
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// Unimplemented.
}

func initHeap() {
	// preinit() may have moved heapStart; reset heapptr
	ptr := heapStart
	if GOARCH == "wasm" {
		// llvm11 and llvm12 do not correctly align the heap on wasm
		ptr = align(ptr)
	}
	heapptr = ptr
}

// setHeapEnd sets a new (larger) heapEnd pointer.
func setHeapEnd(newHeapEnd uintptr) {
	// This "heap" is so simple that simply assigning a new value is good
	// enough.
	heapEnd = newHeapEnd
}

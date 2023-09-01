//go:build gc.none

package runtime

// This GC strategy provides no memory allocation at all. It can be useful to
// detect where in a program memory is allocated (via linker errors) or for
// targets that have far too little RAM even for the leaking memory allocator.

import (
	"unsafe"
)

var gcTotalAlloc uint64 // for runtime.MemStats
var gcMallocs uint64
var gcFrees uint64

func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer

func realloc(ptr unsafe.Pointer, size uintptr) unsafe.Pointer

func free(ptr unsafe.Pointer) {
	// Nothing to free when nothing gets allocated.
}

func GC() {
	// Unimplemented.
}

func SetFinalizer(obj interface{}, finalizer interface{}) {
	// Unimplemented.
}

func initHeap() {
	// Nothing to initialize.
}

func setHeapEnd(newHeapEnd uintptr) {
	// Nothing to do here, this function is never actually called.
}

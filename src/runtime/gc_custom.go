//go:build gc.custom
// +build gc.custom

package runtime

// This GC strategy allows an external GC to be plugged in instead of the builtin
// implementations.
//
// The custom implementation must provide the following functions in the runtime package
// using the go:linkname directive:
//
// - func initHeap()
// - func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer
// - func free(ptr unsafe.Pointer)
// - func GC()
// - func KeepAlive(x interface{})
// - func SetFinalizer(obj interface{}, finalizer interface{})
//
// In addition, if targeting wasi, the following functions should be exported for interoperability
// with wasi libraries that use them. Note, this requires the export directive, not go:linkname.
//
// - func malloc(size uintptr) unsafe.Pointer
// - func free(ptr unsafe.Pointer)
// - func calloc(nmemb, size uintptr) unsafe.Pointer
// - func realloc(oldPtr unsafe.Pointer, size uintptr) unsafe.Pointer

import (
	"unsafe"
)

func initHeap()

func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer

func free(ptr unsafe.Pointer)

func GC()

func KeepAlive(x interface{})

func SetFinalizer(obj interface{}, finalizer interface{})

func setHeapEnd(newHeapEnd uintptr) {
	// Heap is in custom GC so ignore for when called from wasm initialization.
}

func markRoots(start, end uintptr) {
	// GC manages roots so ignore to allow gc_stack_portable to compile
}

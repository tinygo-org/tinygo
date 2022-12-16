//go:build gc.custom
// +build gc.custom

package runtime

// This GC strategy allows an external GC to be plugged in instead of the builtin
// implementations.
//
// The interface defined in this file is not stable and can be broken at anytime, even
// across minor versions.
//
// The custom implementation must provide the following functions in the runtime package
// using the go:linkname directive:
//
// - func initHeap()
// - func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer
// - func free(ptr unsafe.Pointer)
// - func markRoots(start, end uintptr)
// - func GC()
// - func KeepAlive(x interface{})
// - func SetFinalizer(obj interface{}, finalizer interface{})
//
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

// initHeap is called when the heap is first initialized at program start.
func initHeap()

// alloc is called to allocate memory. layout is currently not used.
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer

// free is called to explicitly free a previously allocated pointer.
func free(ptr unsafe.Pointer)

// markRoots is called with the start and end addresses to scan for references.
// It is currently only called with the top and bottom of the stack.
func markRoots(start, end uintptr)

// GC is called to explicitly run garbage collection.
func GC()

// KeepAlive is called to ensure the given object is kept alive until this function call.
func KeepAlive(x interface{})

// SetFinalizer is called to set a finalizer method to run when an object is about to be
// cleaned.
func SetFinalizer(obj interface{}, finalizer interface{})

func setHeapEnd(newHeapEnd uintptr) {
	// Heap is in custom GC so ignore for when called from wasm initialization.
}

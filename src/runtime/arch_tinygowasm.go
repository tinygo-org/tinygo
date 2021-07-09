// +build tinygo.wasm

package runtime

import (
	"unsafe"
)

const GOARCH = "wasm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

//go:extern __heap_base
var heapStartSymbol [0]byte

//go:extern __global_base
var globalsStartSymbol [0]byte

//export llvm.wasm.memory.size.i32
func wasm_memory_size(index int32) int32

//export llvm.wasm.memory.grow.i32
func wasm_memory_grow(index int32, delta int32) int32

var (
	heapStart    = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd      = uintptr(wasm_memory_size(0) * wasmPageSize)
	globalsStart = uintptr(unsafe.Pointer(&globalsStartSymbol))
	globalsEnd   = uintptr(unsafe.Pointer(&heapStartSymbol))
)

const wasmPageSize = 64 * 1024

func align(ptr uintptr) uintptr {
	// Align to 16, which is the alignment of max_align_t:
	// https://godbolt.org/z/dYqTsWrGq
	const heapAlign = 16
	return (ptr + heapAlign - 1) &^ (heapAlign - 1)
}

func getCurrentStackPointer() uintptr

// growHeap tries to grow the heap size. It returns true if it succeeds, false
// otherwise.
func growHeap() bool {
	// Grow memory by the available size, which means the heap size is doubled.
	memorySize := wasm_memory_size(0)
	result := wasm_memory_grow(0, memorySize)
	if result == -1 {
		// Grow failed.
		return false
	}

	setHeapEnd(uintptr(wasm_memory_size(0) * wasmPageSize))

	// Heap has grown successfully.
	return true
}

// The below functions override the default allocator of wasi-libc.
// Most functions are defined but unimplemented to make sure that if there is
// any code using them, they will get an error instead of (incorrectly) using
// the wasi-libc dlmalloc heap implementation instead. If they are needed by any
// program, they can certainly be implemented.

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	return alloc(size, nil)
}

//export free
func libc_free(ptr unsafe.Pointer) {
	free(ptr)
}

//export calloc
func libc_calloc(nmemb, size uintptr) unsafe.Pointer {
	// Note: we could be even more correct here and check that nmemb * size
	// doesn't overflow. However the current implementation should normally work
	// fine.
	return alloc(nmemb*size, nil)
}

//export realloc
func libc_realloc(ptr unsafe.Pointer, size uintptr) unsafe.Pointer {
	runtimePanic("unimplemented: realloc")
	return nil
}

//export posix_memalign
func libc_posix_memalign(memptr *unsafe.Pointer, alignment, size uintptr) int {
	runtimePanic("unimplemented: posix_memalign")
	return 0
}

//export aligned_alloc
func libc_aligned_alloc(alignment, bytes uintptr) unsafe.Pointer {
	runtimePanic("unimplemented: aligned_alloc")
	return nil
}

//export malloc_usable_size
func libc_malloc_usable_size(ptr unsafe.Pointer) uintptr {
	runtimePanic("unimplemented: malloc_usable_size")
	return 0
}

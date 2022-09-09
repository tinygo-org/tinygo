//go:build tinygo.wasm
// +build tinygo.wasm

package runtime

import (
	"unsafe"
)

const GOARCH = "wasm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

const deferExtraRegs = 0

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

// The below functions override the default allocator of wasi-libc. This ensures
// code linked from other languages can allocate memory without colliding with
// our GC allocations.

var allocs = make(map[uintptr][]byte)

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	buf := make([]byte, size)
	ptr := unsafe.Pointer(&buf[0])
	allocs[uintptr(ptr)] = buf
	return ptr
}

//export free
func libc_free(ptr unsafe.Pointer) {
	delete(allocs, uintptr(ptr))
}

//export calloc
func libc_calloc(nmemb, size uintptr) unsafe.Pointer {
	// No difference between calloc and malloc.
	return libc_malloc(nmemb * size)
}

//export realloc
func libc_realloc(oldPtr unsafe.Pointer, size uintptr) unsafe.Pointer {
	// It's hard to optimize this to expand the current buffer with our GC, but
	// it is theoretically possible. For now, just always allocate fresh.
	buf := make([]byte, size)

	if oldPtr != nil {
		oldBuf := allocs[uintptr(oldPtr)]
		delete(allocs, uintptr(oldPtr))
		copy(buf, oldBuf)
	}

	ptr := unsafe.Pointer(&buf[0])
	allocs[uintptr(ptr)] = buf
	return ptr
}

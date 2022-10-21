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

const (
	// wasmMemoryIndex is always zero until the multi-memory feature is used.
	//
	// See https://github.com/WebAssembly/multi-memory
	wasmMemoryIndex = 0

	// wasmPageSize is the size of a page in WebAssembly's 32-bit memory. This
	// is also its only unit of change.
	//
	// See https://www.w3.org/TR/wasm-core-1/#page-size
	wasmPageSize = 64 * 1024
)

// wasm_memory_size invokes the "memory.size" instruction, which returns the
// current size to the memory at the given index (always wasmMemoryIndex), in
// pages.
//
//export llvm.wasm.memory.size.i32
func wasm_memory_size(index int32) int32

// wasm_memory_grow invokes the "memory.grow" instruction, which attempts to
// increase the size of the memory at the given index (always wasmMemoryIndex),
// by the delta (in pages). This returns the previous size on success of -1 on
// failure.
//
//export llvm.wasm.memory.grow.i32
func wasm_memory_grow(index int32, delta int32) int32

var (
	// heapStart is the current memory offset which starts the heap. The heap
	// extends from this offset until heapEnd (exclusive).
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))

	// heapEnd is the current memory length in bytes.
	heapEnd = uintptr(wasm_memory_size(wasmMemoryIndex) * wasmPageSize)

	globalsStart = uintptr(unsafe.Pointer(&globalsStartSymbol))
	globalsEnd   = uintptr(unsafe.Pointer(&heapStartSymbol))
)

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
	memorySize := wasm_memory_size(wasmMemoryIndex)
	result := wasm_memory_grow(wasmMemoryIndex, memorySize)
	if result == -1 {
		// Grow failed.
		return false
	}

	setHeapEnd(uintptr(wasm_memory_size(wasmMemoryIndex) * wasmPageSize))

	// Heap has grown successfully.
	return true
}

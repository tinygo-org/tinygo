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

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
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

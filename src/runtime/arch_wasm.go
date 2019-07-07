// +build wasm

package runtime

import (
	"unsafe"
)

const GOARCH = "wasm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

//go:extern __heap_base
var heapStartSymbol unsafe.Pointer

//go:export llvm.wasm.memory.size.i32
func wasm_memory_size(index int32) int32

var (
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd   = uintptr(wasm_memory_size(0) * wasmPageSize)
)

const wasmPageSize = 64 * 1024

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}

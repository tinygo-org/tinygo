// +build tinygo.arm

package runtime

import (
	"unsafe"
)

const GOARCH = "arm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

//go:extern _heap_start
var heapStartSymbol unsafe.Pointer

//go:extern _heap_end
var heapEndSymbol unsafe.Pointer

var (
	heapStart = uintptr(unsafe.Pointer(&heapStartSymbol))
	heapEnd   = uintptr(unsafe.Pointer(&heapEndSymbol))
)

// Align on word boundary.
func align(ptr uintptr) uintptr {
	return (ptr + 3) &^ 3
}

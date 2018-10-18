// +build wasm,!arm,!avr

package runtime

import (
	"unsafe"
)

const GOARCH = "wasm"

// The length type used inside strings and slices.
type lenType uint32

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

//go:extern __heap_base
var heapStart unsafe.Pointer

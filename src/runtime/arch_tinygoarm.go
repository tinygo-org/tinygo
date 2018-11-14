// +build tinygo.arm

package runtime

import (
	"unsafe"
)

const GOARCH = "arm"

// The bitness of the CPU (e.g. 8, 32, 64).
const TargetBits = 32

//go:extern _heap_start
var heapStart unsafe.Pointer

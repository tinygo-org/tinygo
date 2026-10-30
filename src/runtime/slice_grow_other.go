//go:build !gc.boehm

package runtime

import (
	"math/bits"
	"unsafe"
)

// sliceGrowCap is the capacity for a slice that needs newCap elements.
//
// This can be made more memory-efficient by multiplying by some other constant, such as 1.5,
// which seems to be allowed by the Go language specification (but this can be observed by
// programs); however, due to memory fragmentation and the current state of the TinyGo
// memory allocators, this causes some difficult to debug issues.
func sliceGrowCap(oldCap, newCap, elemSize uintptr, layout unsafe.Pointer) uintptr {
	return 1 << bits.Len(uint(newCap))
}

func sliceGrowFill(buf unsafe.Pointer, newCap, elemSize uintptr, layout unsafe.Pointer) uintptr {
	return newCap
}

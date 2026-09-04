//go:build gc.leaking || gc.none

package runtime

import (
	"internal/gclayout"
	"unsafe"
)

func allocManual(size uintptr) unsafe.Pointer {
	if size == 0 {
		return alloc_zero(size, gclayout.NoPtrs.AsPtr())
	}
	return alloc(size, gclayout.NoPtrs.AsPtr())
}

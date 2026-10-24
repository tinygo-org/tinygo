//go:build scheduler.asyncify && !gc.leaking

package task

import (
	"internal/gclayout"
	"unsafe"
)

// allocStack allocates the stack of a new goroutine.
func allocStack(size uintptr) unsafe.Pointer {
	return runtime_alloc(size, gclayout.Conservative.AsPtr())
}

// releaseStack is called with the stack of a goroutine that has exited.
func releaseStack(base, size uintptr) {
	runtime_freeTaskStack(base)
}

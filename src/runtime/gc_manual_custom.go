//go:build gc.custom

package runtime

import (
	"internal/gclayout"
	"internal/task"
	"unsafe"
)

// Custom collectors retain manual allocations through ordinary typed roots so
// the custom GC interface does not need an additional allocation primitive.
var manualAllocs = make(map[*byte]struct{})
var manualAllocsLock task.PMutex

func allocManual(size uintptr) unsafe.Pointer {
	if size == 0 {
		return alloc_zero(size, gclayout.NoPtrs.AsPtr())
	}
	ptr := alloc(size, gclayout.NoPtrs.AsPtr())
	manualAllocsLock.Lock()
	manualAllocs[(*byte)(ptr)] = struct{}{}
	manualAllocsLock.Unlock()
	return ptr
}

func freeManual(ptr unsafe.Pointer) {
	if ptr == nil || ptr == unsafe.Pointer(zeroSizeAllocPtr) {
		return
	}
	manualAllocsLock.Lock()
	delete(manualAllocs, (*byte)(ptr))
	manualAllocsLock.Unlock()
	free(ptr)
}

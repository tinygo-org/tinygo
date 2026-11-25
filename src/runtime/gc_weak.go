//go:build gc.boehm || gc.conservative || gc.precise

package runtime

import (
	"internal/task"
	"unsafe"
)

type weakHandle struct {
	// ptr stores ^uintptr(target) so conservative GC does not scan it as a
	// live pointer, or 0 once the target becomes unreachable.
	ptr  uintptr
	next *weakHandle
}

var (
	weakLock           task.PMutex
	weakHandles        map[uintptr]*weakHandle
	weakPruneThreshold = 64
)

//go:linkname registerWeakPointer weak.runtime_registerWeakPointer
func registerWeakPointer(ptr unsafe.Pointer) unsafe.Pointer {
	enc := ^uintptr(ptr)
	weakLock.Lock()
	if h := weakHandles[enc]; h != nil {
		gcLock.Lock()
		alive := h.ptr != 0
		gcLock.Unlock()
		if alive {
			weakLock.Unlock()
			KeepAlive(ptr)
			return unsafe.Pointer(h)
		}
	}
	h := &weakHandle{ptr: enc}
	registerWeakHandle(h, ptr)
	if weakHandles == nil {
		weakHandles = make(map[uintptr]*weakHandle)
	} else if len(weakHandles) >= weakPruneThreshold {
		for k, entry := range weakHandles {
			gcLock.Lock()
			dead := entry.ptr == 0
			gcLock.Unlock()
			if dead {
				delete(weakHandles, k)
			}
		}
		if len(weakHandles)*2 > weakPruneThreshold {
			weakPruneThreshold = len(weakHandles) * 2
		} else if weakPruneThreshold > 64 && len(weakHandles)*4 < weakPruneThreshold {
			weakPruneThreshold /= 2
		}
	}
	weakHandles[enc] = h
	weakLock.Unlock()
	KeepAlive(ptr)
	return unsafe.Pointer(h)
}

//go:linkname makeStrongFromWeak weak.runtime_makeStrongFromWeak
func makeStrongFromWeak(u unsafe.Pointer) unsafe.Pointer {
	gcLock.Lock()
	enc := (*weakHandle)(u).ptr
	gcLock.Unlock()
	if enc == 0 {
		return nil
	}
	return unsafe.Pointer(^enc)
}

//go:build tinygo.wasm && !(custommalloc || wasm_unknown || gc.boehm)

package runtime

import (
	"internal/task"
	"unsafe"
)

// The below functions override the default allocator of wasi-libc. This ensures
// code linked from other languages can allocate memory without colliding with
// our GC allocations.

// Map of allocations, where the key is the allocation address and the value is
// its size. Integer keys intentionally do not act as GC roots: manual
// allocations are retained by the allocator until free.
var allocs = make(map[uintptr]uintptr)
var allocsLock task.PMutex

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	if size == 0 {
		return nil
	}
	ptr := allocManual(size)
	allocsLock.Lock()
	allocs[uintptr(ptr)] = size
	allocsLock.Unlock()
	return ptr
}

//export free
func libc_free(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}
	allocsLock.Lock()
	if _, ok := allocs[uintptr(ptr)]; ok {
		delete(allocs, uintptr(ptr))
		allocsLock.Unlock()
		freeManual(ptr)
	} else {
		allocsLock.Unlock()
		runtimeFatal("free: invalid pointer")
	}
}

//export calloc
func libc_calloc(nmemb, size uintptr) unsafe.Pointer {
	if size != 0 && nmemb > ^uintptr(0)/size {
		return nil
	}
	return libc_malloc(nmemb * size)
}

//export realloc
func libc_realloc(oldPtr unsafe.Pointer, size uintptr) unsafe.Pointer {
	if size == 0 {
		libc_free(oldPtr)
		return nil
	}

	// It's hard to optimize this to expand the current buffer with our GC, but
	// it is theoretically possible. For now, just always allocate fresh.
	// TODO: we could skip this if the new allocation is smaller than the old.
	ptr := allocManual(size)

	allocsLock.Lock()
	if oldPtr != nil {
		if oldSize, ok := allocs[uintptr(oldPtr)]; ok {
			oldBuf := unsafe.Slice((*byte)(oldPtr), oldSize)
			newBuf := unsafe.Slice((*byte)(ptr), size)
			copy(newBuf, oldBuf)
			delete(allocs, uintptr(oldPtr))
		} else {
			allocsLock.Unlock()
			runtimeFatal("realloc: invalid pointer")
		}
	}
	allocs[uintptr(ptr)] = size
	allocsLock.Unlock()
	if oldPtr != nil {
		freeManual(oldPtr)
	}
	return ptr
}

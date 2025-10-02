//go:build tinygo.wasm && !(custommalloc || wasm_unknown || gc.boehm)

package runtime

import "unsafe"

// The below functions override the default allocator of wasi-libc. This ensures
// code linked from other languages can allocate memory without colliding with
// our GC allocations.

// Map of allocations, where the key is the allocated pointer and the value is
// the size of the allocation.
var allocs = make(map[unsafe.Pointer]uintptr)

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	if size == 0 {
		return nil
	}
	const wordSize = unsafe.Sizeof(unsafe.Pointer(nil))
	buf := make([]unsafe.Pointer, (size+wordSize-1)/wordSize)
	ptr := unsafe.Pointer(&buf[0])
	allocs[ptr] = size
	return ptr
}

//export free
func libc_free(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}
	if _, ok := allocs[ptr]; ok {
		delete(allocs, ptr)
	} else {
		panic("free: invalid pointer")
	}
}

//export calloc
func libc_calloc(nmemb, size uintptr) unsafe.Pointer {
	// No difference between calloc and malloc.
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
	const wordSize = unsafe.Sizeof(unsafe.Pointer(nil))
	buf := make([]unsafe.Pointer, (size+wordSize-1)/wordSize)
	ptr := unsafe.Pointer(&buf[0])

	if oldPtr != nil {
		if oldSize, ok := allocs[oldPtr]; ok {
			oldBuf := unsafe.Slice((*byte)(oldPtr), oldSize)
			newBuf := unsafe.Slice((*byte)(ptr), size)
			copy(newBuf, oldBuf)
			delete(allocs, oldPtr)
		} else {
			panic("realloc: invalid pointer")
		}
	}

	allocs[ptr] = size
	return ptr
}

//go:build tinygo.wasm && !custommalloc

package runtime

import "unsafe"

// The below functions override the default allocator of wasi-libc. This ensures
// code linked from other languages can allocate memory without colliding with
// our GC allocations.

var allocs = make(map[uintptr][]byte)

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	if size == 0 {
		size = 1 // Match dlmalloc's behavior to allocate a "minimum-sized chunk"
	}
	buf := make([]byte, size)
	ptr := unsafe.Pointer(&buf[0])
	allocs[uintptr(ptr)] = buf
	return ptr
}

//export free
func libc_free(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}
	if _, ok := allocs[uintptr(ptr)]; ok {
		delete(allocs, uintptr(ptr))
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
		size = 1 // Match dlmalloc's behavior to allocate a "minimum-sized chunk"
	}

	// It's hard to optimize this to expand the current buffer with our GC, but
	// it is theoretically possible. For now, just always allocate fresh.
	buf := make([]byte, size)

	if oldPtr != nil {
		if oldBuf, ok := allocs[uintptr(oldPtr)]; ok {
			copy(buf, oldBuf)
			delete(allocs, uintptr(oldPtr))
		} else {
			panic("realloc: invalid pointer")
		}
	}

	ptr := unsafe.Pointer(&buf[0])
	allocs[uintptr(ptr)] = buf
	return ptr
}

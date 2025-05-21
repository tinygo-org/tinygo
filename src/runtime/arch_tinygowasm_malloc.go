//go:build tinygo.wasm && !(custommalloc || wasm_unknown || gc.boehm)

package runtime

import "unsafe"

// The below functions override the default allocator of wasi-libc. This ensures
// code linked from other languages can allocate memory without colliding with
// our GC allocations.

// Map of allocations, where the key is the allocated pointer and the value is
// the size of the allocation.
// TODO: make this a map[unsafe.Pointer]uintptr, since that results in slightly
// smaller binaries. But for that to work, unsafe.Pointer needs to be seen as a
// binary key (which it is not at the moment).
// See https://github.com/tinygo-org/tinygo/pull/4898 for details.
var allocs = make(map[*byte]uintptr)

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	if size == 0 {
		return nil
	}
	ptr := alloc(size, nil)
	allocs[(*byte)(ptr)] = size
	return ptr
}

//export free
func libc_free(ptr unsafe.Pointer) {
	if ptr == nil {
		return
	}
	if _, ok := allocs[(*byte)(ptr)]; ok {
		delete(allocs, (*byte)(ptr))
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
	ptr := alloc(size, nil)

	if oldPtr != nil {
		if oldSize, ok := allocs[(*byte)(oldPtr)]; ok {
			oldBuf := unsafe.Slice((*byte)(oldPtr), oldSize)
			newBuf := unsafe.Slice((*byte)(ptr), size)
			copy(newBuf, oldBuf)
			delete(allocs, (*byte)(oldPtr))
		} else {
			panic("realloc: invalid pointer")
		}
	}

	allocs[(*byte)(ptr)] = size
	return ptr
}

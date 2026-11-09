//go:build wasip3 && !custommalloc

package runtime

import "unsafe"

// wasi-libc allocates thread-local storage with posix_memalign.
//
//export posix_memalign
func libc_posix_memalign(memptr *unsafe.Pointer, alignment, size uintptr) int32 {
	// The manual allocator returns 16-byte aligned memory.
	if alignment > 16 {
		return 28 // EINVAL
	}
	ptr := libc_malloc(size)
	if ptr == nil && size != 0 {
		return 48
	}
	*memptr = ptr
	return 0
}

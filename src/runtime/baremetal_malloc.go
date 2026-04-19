//go:build baremetal && !buildmode.c_archive

package runtime

import "unsafe"

//export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	// Note: this zeroes the returned buffer which is not necessary.
	// The same goes for bytealg.MakeNoZero.
	return alloc(size, nil)
}

//export calloc
func libc_calloc(nmemb, size uintptr) unsafe.Pointer {
	// No difference between calloc and malloc.
	return libc_malloc(nmemb * size)
}

//export free
func libc_free(ptr unsafe.Pointer) {
	free(ptr)
}

// +build baremetal

package runtime

import "unsafe"

//go:export malloc
func libc_malloc(size uintptr) unsafe.Pointer {
	return alloc(size)
}

//go:export free
func libc_free(ptr unsafe.Pointer) {
	free(ptr)
}

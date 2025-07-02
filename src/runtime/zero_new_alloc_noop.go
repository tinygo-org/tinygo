//go:build gc.leaking && !baremetal && !nintendoswitch

package runtime

import (
	"unsafe"
)

//go:inline
func zero_new_alloc(ptr unsafe.Pointer, size uintptr) {
	// Wasm linear memory is initialized to zero by default, so
	// there's no need to do anything.  This is also the case for
	// fresh-allocated memory from an mmap() system call.
}

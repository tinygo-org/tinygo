//go:build gc.leaking && (baremetal || nintendoswitch)

package runtime

import (
	"unsafe"
)

//go:inline
func zero_new_alloc(ptr unsafe.Pointer, size uintptr) {
	memzero(ptr, size)
}

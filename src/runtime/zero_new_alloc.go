//go:build gc.leaking && (baremetal || nintendoswitch)

package runtime

import (
	"unsafe"
)

func zero_new_alloc(ptr unsafe.Pointer, size uintptr) {
	memzero(ptr, size)
}

// +build gc.list,!avr

package runtime

import "unsafe"

// markMem scans a memory region for pointers and marks anything that is pointed to.
func markMem(start, end uintptr) {
	start = align(start)
	end &^= unsafe.Alignof(unsafe.Pointer(nil)) - 1
	for ; start < end; start += unsafe.Alignof(unsafe.Pointer(nil)) {
		markAddr(*(*uintptr)(unsafe.Pointer(start)))
	}
}

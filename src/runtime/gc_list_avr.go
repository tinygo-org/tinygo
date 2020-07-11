// +build gc.list,avr

package runtime

import "unsafe"

// markMem scans a memory region for pointers and marks anything that is pointed to.
func markMem(start, end uintptr) {
	if start >= end {
		return
	}
	prevByte := *(*byte)(unsafe.Pointer(start))
	for ; start != end; start++ {
		b := *(*byte)(unsafe.Pointer(start))
		addr := (uintptr(b) << 8) | uintptr(prevByte)
		markAddr(addr)
		prevByte = b
	}
}

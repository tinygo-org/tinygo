// +build tinygo.riscv

package runtime

import "unsafe"

// Implement memset for LLVM.
//go:export memset
func libc_memset(ptr unsafe.Pointer, c byte, size uintptr) {
	for i := uintptr(0); i < size; i++ {
		*(*byte)(unsafe.Pointer(uintptr(ptr) + i)) = c
	}
}

// Implement memmove for LLVM.
//go:export memmove
func libc_memmove(dst, src unsafe.Pointer, size uintptr) {
	memmove(dst, src, size)
}

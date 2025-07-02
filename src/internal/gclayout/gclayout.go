package gclayout

import "unsafe"

// Internal constants for gc layout
// See runtime/gc_precise.go

var (
	NoPtrs  unsafe.Pointer
	Pointer unsafe.Pointer
	String  unsafe.Pointer
	Slice   unsafe.Pointer
)

func init() {
	var sizeBits uintptr

	switch unsafe.Sizeof(uintptr(0)) {
	case 8:
		sizeBits = 6
	case 4:
		sizeBits = 5
	case 2:
		sizeBits = 4
	}

	var sizeShift = sizeBits + 1

	NoPtrs = unsafe.Pointer(uintptr(0b0<<sizeShift) | uintptr(0b1<<1) | uintptr(1))
	Pointer = unsafe.Pointer(uintptr(0b1<<sizeShift) | uintptr(0b1<<1) | uintptr(1))
	String = unsafe.Pointer(uintptr(0b01<<sizeShift) | uintptr(0b10<<1) | uintptr(1))
	Slice = unsafe.Pointer(uintptr(0b001<<sizeShift) | uintptr(0b11<<1) | uintptr(1))
}

package gclayout

import "unsafe"

// Internal constants for gc layout
// See runtime/gc_precise.go

type Layout uintptr

const (
	// 16-bit int => bits = 4
	// 32-bit int => bits = 5
	// 64-bit int => bits = 6
	sizeBits = 4 + unsafe.Sizeof(uintptr(0))/4

	sizeShift = sizeBits + 1

	NoPtrs  = Layout(uintptr(0b0<<sizeShift) | uintptr(0b1<<1) | uintptr(1))
	Pointer = Layout(uintptr(0b1<<sizeShift) | uintptr(0b1<<1) | uintptr(1))
	String  = Layout(uintptr(0b01<<sizeShift) | uintptr(0b10<<1) | uintptr(1))
	Slice   = Layout(uintptr(0b001<<sizeShift) | uintptr(0b11<<1) | uintptr(1))
)

func (l Layout) AsPtr() unsafe.Pointer { return unsafe.Pointer(l) }

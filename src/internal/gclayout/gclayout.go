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

	ptrAlign = unsafe.Alignof(uintptr(0))

	sizeShift = sizeBits + 1

	NoPtrs  = Layout((0 << sizeShift) | (1 << 1) | 1)
	Pointer = Layout((1 << sizeShift) | ((unsafe.Sizeof(unsafe.Pointer(nil)) / ptrAlign) << 1) | 1)
	String  = Layout((1 << sizeShift) | ((unsafe.Sizeof("") / ptrAlign) << 1) | 1)
	Slice   = Layout((1 << sizeShift) | ((unsafe.Sizeof([]byte{}) / ptrAlign) << 1) | 1)
)

func (l Layout) AsPtr() unsafe.Pointer { return unsafe.Pointer(l) }

package main

// This file tests various operations on pointers, such as pointer arithmetic
// and dereferencing pointers.

import "unsafe"

// Dereference pointers.

func pointerDerefZero(x *[0]int) [0]int {
	return *x // This is a no-op, there is nothing to load.
}

// Unsafe pointer casts, they are sometimes a no-op.

func pointerCastFromUnsafe(x unsafe.Pointer) *int {
	return (*int)(x)
}

func pointerCastToUnsafe(x *int) unsafe.Pointer {
	return unsafe.Pointer(x)
}

func pointerCastToUnsafeNoop(x *byte) unsafe.Pointer {
	return unsafe.Pointer(x)
}

// The compiler has support for a few special cast+add patterns that are
// transformed into a single GEP.

func pointerUnsafeGEPFixedOffset(ptr *byte) *byte {
	return (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + 10))
}

func pointerUnsafeGEPByteOffset(ptr *byte, offset uintptr) *byte {
	return (*byte)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + offset))
}

func pointerUnsafeGEPIntOffset(ptr *int32, offset uintptr) *int32 {
	return (*int32)(unsafe.Pointer(uintptr(unsafe.Pointer(ptr)) + offset*4))
}

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

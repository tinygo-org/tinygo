package abi

import "unsafe"

// Tell the compiler the given pointer doesn't escape.
// The compiler knows about this function and will give the nocapture parameter
// attribute.
func NoEscape(p unsafe.Pointer) unsafe.Pointer {
	return p
}

func Escape[T any](x T) T {
	// This function is either implemented in the compiler, or left undefined
	// for some variation of T. The body of this function should not be compiled
	// as-is.
	panic("internal/abi.Escape: unreachable (implemented in the compiler)")
}

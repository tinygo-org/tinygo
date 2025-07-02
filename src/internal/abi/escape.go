package abi

import "unsafe"

// Tell the compiler the given pointer doesn't escape.
// The compiler knows about this function and will give the nocapture parameter
// attribute.
func NoEscape(p unsafe.Pointer) unsafe.Pointer {
	return p
}

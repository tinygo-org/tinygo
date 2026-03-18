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

// EscapeNonString forces v to be on the heap, if v contains a
// non-string pointer.
//
// This is used in hash/maphash.Comparable. We cannot hash pointers
// to local variables on stack, as their addresses might change on
// stack growth. Strings are okay as the hash depends on only the
// content, not the pointer.
//
// This is essentially
//
//	if hasNonStringPointers(T) { Escape(v) }
//
// Implemented as a compiler intrinsic.
func EscapeNonString[T any](v T) { panic("intrinsic") }

// EscapeToResultNonString models a data flow edge from v to the result,
// if v contains a non-string pointer. If v contains only string pointers,
// it returns a copy of v, but is not modeled as a data flow edge
// from the escape analysis's perspective.
//
// This is used in unique.clone, to model the data flow edge on the
// value with strings excluded, because strings are cloned (by
// content).
//
// TODO: probably we should define this as a intrinsic and EscapeNonString
// could just be "heap = EscapeToResultNonString(v)". This way we can model
// an edge to the result but not necessarily heap.
func EscapeToResultNonString[T any](v T) T {
	EscapeNonString(v)
	return *(*T)(NoEscape(unsafe.Pointer(&v)))
}

package runtime

// Shared code for the various garbage collectors.

import "unsafe"

// Special alloc function that should never actually be called.
// It is used instead of normal alloc in //go:noheap functions, and must either
// be optimized away or throw a linker error.
func alloc_noheap(size uintptr, layout unsafe.Pointer) unsafe.Pointer

// Special alloc function that returns a sentinel value that can never be on the
// heap or match any other valid pointer. An alloc(0, xxx) call can be safely
// converted to an alloc_zero(0, xxx) call as an optimization.
//
// It is always a good idea to inline this function, since the result is a
// constant. Marking it as go:inline to be sure even though the compiler should
// already be doing this.
//
//go:inline
func alloc_zero(size uintptr, layout unsafe.Pointer) unsafe.Pointer {
	// Returning a constant here is safe, since the Go spec does not require
	// multiple zero-sized allocations to be unequal when compared for equality:
	//
	// > Pointers to distinct zero-size variables may or may not be equal.
	//
	// Source: https://go.dev/ref/spec#Comparison_operators
	return unsafe.Pointer(zeroSizeAllocPtr)
}

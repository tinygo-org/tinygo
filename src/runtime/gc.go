package runtime

// Shared code for the various garbage collectors.

import "unsafe"

// Special alloc function that should never actually be called.
// It is used instead of normal alloc in //go:noheap functions, and must either
// be optimized away or throw a linker error.
func alloc_noheap(size uintptr, layout unsafe.Pointer) unsafe.Pointer

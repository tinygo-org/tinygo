//go:build scheduler.asyncify && gc.leaking

package task

import "unsafe"

// The leaking GC never frees memory, so keep the stacks of exited goroutines
// for reuse. Each free stack stores its list entry in its own first words.
type freeStack struct {
	next *freeStack
	size uintptr
}

var freeStacks *freeStack

// allocStack allocates the stack of a new goroutine, reusing and zeroing the
// stack of an exited goroutine of the same size if there is one.
func allocStack(size uintptr) unsafe.Pointer {
	// Every goroutine gets the same stack size with Asyncify, so the first free
	// stack always matches and this loop does not search.
	for p := &freeStacks; *p != nil; p = &(*p).next {
		if s := *p; s.size == size {
			*p = s.next
			clear(unsafe.Slice((*byte)(unsafe.Pointer(s)), size))
			return unsafe.Pointer(s)
		}
	}
	return runtime_alloc(size, stackLayout())
}

// releaseStack is called with the stack of a goroutine that has exited. The
// goroutine no longer runs on it, so it can be reused.
func releaseStack(base, size uintptr) {
	s := (*freeStack)(unsafe.Pointer(base))
	*s = freeStack{next: freeStacks, size: size}
	freeStacks = s
}

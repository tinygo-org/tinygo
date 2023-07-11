//go:build gc.conservative

// This implements the block-based heap as a fully conservative GC. No tracking
// of pointers is done, every word in an object is considered live if it looks
// like a pointer.

package runtime

const preciseHeap = false

type gcObjectScanner struct {
}

func newGCObjectScanner(block gcBlock) gcObjectScanner {
	return gcObjectScanner{}
}

func (scanner *gcObjectScanner) pointerFree() bool {
	// We don't know whether this object contains pointers, so conservatively
	// return false.
	return false
}

// nextIsPointer returns whether this could be a pointer. Because the GC is
// conservative, we can't do much more than check whether the object lies
// somewhere in the heap.
func (scanner gcObjectScanner) nextIsPointer(ptr, parent, addrOfWord uintptr) bool {
	return isOnHeap(ptr)
}

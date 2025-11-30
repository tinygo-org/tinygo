//go:build gc.conservative

// This implements the block-based heap as a fully conservative GC. No tracking
// of pointers is done, every word in an object is considered live if it looks
// like a pointer.

package runtime

import "unsafe"

// gcLayout tracks pointer locations in a heap object.
// The conservative GC treats all locations as potential pointers, so this doesn't need to store anything.
type gcLayout struct {
}

// parseGCLayout stores the layout information passed to alloc into a gcLayout value.
// The conservative GC discards this information.
func parseGCLayout(layout unsafe.Pointer) gcLayout {
	return gcLayout{}
}

// scanner creates a gcObjectScanner with this layout.
func (l gcLayout) scanner() gcObjectScanner {
	return gcObjectScanner{}
}

type gcObjectScanner struct {
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

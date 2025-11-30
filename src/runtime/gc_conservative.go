//go:build gc.conservative

// This implements the block-based heap as a fully conservative GC. No tracking
// of pointers is done, every word in an object is considered live if it looks
// like a pointer.

package runtime

import "unsafe"

// parseGCLayout stores the layout information passed to alloc into a gcLayout value.
// The conservative GC discards this information.
func parseGCLayout(layout unsafe.Pointer) gcLayout {
	return gcLayout{}
}

// gcLayout tracks pointer locations in a heap object.
// The conservative GC treats all locations as potential pointers, so this doesn't need to store anything.
type gcLayout struct {
}

func (l gcLayout) pointerFree() bool {
	// We don't know whether this object contains pointers, so conservatively
	// return false.
	return false
}

func (l gcLayout) scan(start, len uintptr) {
	scanConservative(start, len)
}

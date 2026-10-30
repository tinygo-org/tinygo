package runtime

// This file implements compiler builtins for slices: append() and copy().

import (
	"unsafe"
)

// Builtin append(src, elements...) function: append elements to src and return
// the modified (possibly expanded) slice.
func sliceAppend(srcBuf, elemsBuf unsafe.Pointer, srcLen, srcCap, elemsLen, elemSize uintptr, layout unsafe.Pointer) (unsafe.Pointer, uintptr, uintptr) {
	newLen := srcLen + elemsLen
	if elemsLen > 0 {
		// Allocate a new slice with capacity for elemsLen more elements, if necessary;
		// otherwise, reuse the passed slice.
		srcBuf, _, srcCap = sliceGrow(srcBuf, srcLen, srcCap, newLen, elemSize, layout)

		// Append the new elements in-place.
		memmove(unsafe.Add(srcBuf, srcLen*elemSize), elemsBuf, elemsLen*elemSize)
	}

	return srcBuf, newLen, srcCap
}

// sliceGrow returns a new slice with space for at least newCap elements
func sliceGrow(oldBuf unsafe.Pointer, oldLen, oldCap, newCap, elemSize uintptr, layout unsafe.Pointer) (unsafe.Pointer, uintptr, uintptr) {
	if oldCap >= newCap {
		// No need to grow, return the input slice.
		return oldBuf, oldLen, oldCap
	}

	newCap = sliceGrowCap(oldCap, newCap, elemSize, layout)

	buf := alloc(newCap*elemSize, layout)
	newCap = sliceGrowFill(buf, newCap, elemSize, layout)
	if oldLen > 0 {
		// copy any data to new slice
		memmove(buf, oldBuf, oldLen*elemSize)
	}

	return buf, oldLen, newCap
}

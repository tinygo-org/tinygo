package runtime

// This file implements compiler builtins for slices: append() and copy().

import (
	"math/bits"
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
	maxCap := maxSliceCap(elemSize)
	if newCap > maxCap {
		slicePanic()
	}

	// This can be made more memory-efficient by multiplying by some other constant, such as 1.5,
	// which seems to be allowed by the Go language specification (but this can be observed by
	// programs); however, due to memory fragmentation and the current state of the TinyGo
	// memory allocators, this causes some difficult to debug issues.
	if rounded := uintptr(1) << bits.Len(uint(newCap)); rounded > newCap && rounded <= maxCap {
		newCap = rounded
	}

	buf := alloc(newCap*elemSize, layout)
	if oldLen > 0 {
		// copy any data to new slice
		memmove(buf, oldBuf, oldLen*elemSize)
	}

	return buf, oldLen, newCap
}

// maxSliceCap returns the largest capacity that fits in an int and whose size
// in bytes fits in a uintptr.
func maxSliceCap(elemSize uintptr) uintptr {
	maxInt := int(^uint(0) >> 1)
	limit := ^uintptr(0)
	if uint64(maxInt) < uint64(limit) {
		limit = uintptr(maxInt)
	}
	if elemSize > 1 && ^uintptr(0)/elemSize < limit {
		limit = ^uintptr(0) / elemSize
	}
	return limit
}

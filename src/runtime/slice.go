package runtime

// This file implements compiler builtins for slices: append() and copy().

import (
	"math/bits"
	"unsafe"
)

// Builtin append(src, elements...) function: append elements to src and return
// the modified (possibly expanded) slice.
func sliceAppend(srcBuf, elemsBuf unsafe.Pointer, srcLen, srcCap, elemsLen, elemSize uintptr) (unsafe.Pointer, uintptr, uintptr) {
	newLen := srcLen + elemsLen
	if newLen > srcLen {
		// Allocate a new slice with capacity for elemsLen more elements, if necessary;
		// otherwise, reuse the passed slice.
		srcBuf, _, srcCap = sliceGrow(srcBuf, srcLen, srcCap, newLen, elemSize)

		if elemsLen > 0 {
			// Append the new elements in-place.
			memmove(unsafe.Add(srcBuf, srcLen*elemSize), elemsBuf, elemsLen*elemSize)
		}
	}

	return srcBuf, newLen, srcCap
}

// Builtin copy(dst, src) function: copy bytes from dst to src.
func sliceCopy(dst, src unsafe.Pointer, dstLen, srcLen uintptr, elemSize uintptr) int {
	// n = min(srcLen, dstLen)
	n := srcLen
	if n > dstLen {
		n = dstLen
	}
	memmove(dst, src, n*elemSize)
	return int(n)
}

// sliceGrow returns a new slice with space for at least newCap elements
func sliceGrow(oldBuf unsafe.Pointer, oldLen, oldCap, newCap, elemSize uintptr) (unsafe.Pointer, uintptr, uintptr) {
	if oldCap >= newCap {
		// No need to grow, return the input slice.
		return oldBuf, oldLen, oldCap
	}

	// This can be made more memory-efficient by multiplying by some other constant, such as 1.5,
	// which seems to be allowed by the Go language specification (but this can be observed by
	// programs).
	newCap = 1 << (32 - bits.LeadingZeros32(uint32(newCap)))

	buf := alloc(newCap*elemSize, nil)
	if oldLen > 0 {
		// copy any data to new slice
		memmove(buf, oldBuf, oldLen*elemSize)
	}

	return buf, oldLen, newCap
}

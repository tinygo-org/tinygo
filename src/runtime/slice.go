package runtime

// This file implements compiler builtins for slices: append() and copy().

import (
	"unsafe"
)

// Builtin append(src, elements...) function: append elements to src and return
// the modified (possibly expanded) slice.
func sliceAppend(srcBuf, elemsBuf unsafe.Pointer, srcLen, srcCap, elemsLen uintptr, elemSize uintptr) (unsafe.Pointer, uintptr, uintptr) {
	if elemsLen == 0 {
		// Nothing to append, return the input slice.
		return srcBuf, srcLen, srcCap
	}

	if srcLen+elemsLen > srcCap {
		// Slice does not fit, allocate a new buffer that's large enough.
		srcCap = srcCap * 2
		if srcCap == 0 { // e.g. zero slice
			srcCap = 1
		}
		for srcLen+elemsLen > srcCap {
			// This algorithm may be made more memory-efficient: don't multiply
			// by two but by 1.5 or something. As far as I can see, that's
			// allowed by the Go language specification (but may be observed by
			// programs).
			srcCap *= 2
		}
		buf := alloc(srcCap*elemSize, nil)

		// Copy the old slice to the new slice.
		if srcLen != 0 {
			memmove(buf, srcBuf, srcLen*elemSize)
		}
		srcBuf = buf
	}

	// The slice fits (after possibly allocating a new one), append it in-place.
	memmove(unsafe.Add(srcBuf, srcLen*elemSize), elemsBuf, elemsLen*elemSize)
	return srcBuf, srcLen + elemsLen, srcCap
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

	// TODO(dgryski): sliceGrow() and sliceAppend() should be refactored to share the base growth code.

	if oldCap >= newCap {
		// No need to grow, return the input slice.
		return oldBuf, oldLen, oldCap
	}

	// allow nil slice
	if oldCap == 0 {
		oldCap++
	}

	// grow capacity
	for oldCap < newCap {
		oldCap *= 2
	}

	buf := alloc(oldCap*elemSize, nil)
	if oldLen > 0 {
		// copy any data to new slice
		memmove(buf, oldBuf, oldLen*elemSize)
	}

	return buf, oldLen, oldCap
}

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

		// sliceGrow allocates, and allocating may collect. elemsBuf must still
		// be reachable when the memmove above reads it, so it has to be a GC
		// root for the duration of the grow. As always with KeepAlive, this
		// marks the end of the required lifetime: it keeps elemsBuf alive up
		// to this point, which is what covers the sliceGrow call above.
		//
		// Upstream Go does not need this. Its compiler derives liveness and
		// emits a stack map at every safepoint, so anything still live across
		// a call is a root automatically, and KeepAlive is only reached for
		// when a value must outlive what that analysis can see. We have no
		// such map. Roots exist only where the compiler was told to spill a
		// pointer into the stack object, so the obligation is inverted: a
		// pointer that must survive an allocation has to be marked here.
		//
		// Note this relies on spills for parameters being emitted at function
		// entry, so the slot covers the whole body including the grow above.
		// See transform.MakeGCStackSlots.
		keepAlivePointer(elemsBuf)
	}

	return srcBuf, newLen, srcCap
}

// keepAlivePointer makes sure ptr is still considered live at this point, for
// the garbage collector as well as the compiler.
//
// This is a compiler intrinsic.
func keepAlivePointer(ptr unsafe.Pointer)

// sliceGrow returns a new slice with space for at least newCap elements
func sliceGrow(oldBuf unsafe.Pointer, oldLen, oldCap, newCap, elemSize uintptr, layout unsafe.Pointer) (unsafe.Pointer, uintptr, uintptr) {
	if oldCap >= newCap {
		// No need to grow, return the input slice.
		return oldBuf, oldLen, oldCap
	}

	// This can be made more memory-efficient by multiplying by some other constant, such as 1.5,
	// which seems to be allowed by the Go language specification (but this can be observed by
	// programs); however, due to memory fragmentation and the current state of the TinyGo
	// memory allocators, this causes some difficult to debug issues.
	newCap = 1 << bits.Len(uint(newCap))

	buf := alloc(newCap*elemSize, layout)
	if oldLen > 0 {
		// copy any data to new slice
		memmove(buf, oldBuf, oldLen*elemSize)
	}

	return buf, oldLen, newCap
}

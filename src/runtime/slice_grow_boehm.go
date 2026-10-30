//go:build gc.boehm

package runtime

import (
	"internal/gclayout"
	"unsafe"
)

// sliceGrowCap is the capacity for a slice of oldCap that needs newCap, as Go
// grows it: double while small, then by a quarter, see nextslicecap in Go's
// runtime/slice.go. Doubling a large slice left it up to half empty.
func sliceGrowCap(oldCap, newCap, elemSize uintptr, layout unsafe.Pointer) uintptr {
	c := oldCap + oldCap
	const threshold = 256
	switch {
	case newCap > c:
		c = newCap
	case oldCap >= threshold:
		c = oldCap
		for c < newCap {
			next := c + (c+3*threshold)>>2
			if next <= c {
				return newCap
			}
			c = next
		}
	}
	if elemSize == 0 || c > (^uintptr(0)>>2)/elemSize || !sliceUntyped(layout) {
		return c
	}
	return sliceFitCap(c, newCap, elemSize)
}

// sliceUntyped reports whether a slice of layout is allocated without a
// bdwgc type descriptor, which takes extra room at the end of the object.
func sliceUntyped(layout unsafe.Pointer) bool {
	switch layout {
	case gclayout.NoPtrs.AsPtr(), gclayout.Pointer.AsPtr(), gclayout.PointerPair.AsPtr():
		return true
	}
	return false
}

// sliceFitCap picks a capacity near c that fills a whole bdwgc object: one 16
// byte granule for a tiny one, whole 4096 byte blocks for one over 2048 bytes.
// Fit in the smaller object when that loses at most an eighth of c.
func sliceFitCap(c, newCap, elemSize uintptr) uintptr {
	size := c * elemSize
	var unit uintptr
	switch {
	case size < 384:
		// bdwgc's tiny size classes, one per 16 byte granule.
		unit = 16
	case size > 2048:
		// Large objects take whole 4096 byte blocks.
		unit = 4096
	default:
		// Coarser size classes that bdwgc picks at run time, see
		// sliceGrowFill.
		return c
	}
	if fit := size &^ (unit - 1) / elemSize; fit >= newCap && fit >= c-c/8 {
		return fit
	}
	return (size + unit - 1) &^ (unit - 1) / elemSize
}

// sliceGrowFill extends the capacity of a new small slice buffer to the size
// class bdwgc gave it, zeroing the extra elements. Typed objects keep their
// descriptor in the last word and are scanned only up to the requested size,
// so only the untyped kinds can grow into the tail.
func sliceGrowFill(buf unsafe.Pointer, newCap, elemSize uintptr, layout unsafe.Pointer) uintptr {
	size := newCap * elemSize
	if size == 0 || size > 2048 || !sliceUntyped(layout) {
		return newCap
	}
	usable := libgc_size(uintptr(buf)) / elemSize
	if usable <= newCap {
		return newCap
	}
	memzero(unsafe.Add(buf, size), (usable-newCap)*elemSize)
	return usable
}

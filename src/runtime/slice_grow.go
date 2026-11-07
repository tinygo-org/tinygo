//go:build !baremetal && !gc.leaking

package runtime

// nextSliceCap returns the capacity for a slice of oldCap that needs newLen
// elements, as nextslicecap in Go's runtime/slice.go does.
func nextSliceCap(newLen, oldCap uintptr) uintptr {
	doubleCap := oldCap + oldCap
	if newLen > doubleCap {
		return newLen
	}
	const threshold = 256
	if oldCap < threshold {
		return doubleCap
	}
	newCap := oldCap
	for newCap < newLen {
		next := newCap + (newCap+3*threshold)>>2
		if next < newCap {
			return newLen
		}
		newCap = next
	}
	return newCap
}

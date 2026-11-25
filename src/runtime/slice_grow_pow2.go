//go:build baremetal || gc.leaking

package runtime

import "math/bits"

// nextSliceCap returns the next power of two above newLen. Slower growth
// fragments a small heap, and the leaking collector keeps every old copy,
// so both do better with doubling.
func nextSliceCap(newLen, oldCap uintptr) uintptr {
	return 1 << bits.Len(uint(newLen))
}

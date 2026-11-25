//go:build !mips

package reflectlite

import "unsafe"

const isBigEndian = false

// loadValue loads a value that may or may not be word-aligned. The number of
// bytes given in size are loaded. The biggest possible size it can load is that
// of an uintptr.
func loadValue(ptr unsafe.Pointer, size uintptr) uintptr {
	loadedValue := uintptr(0)
	shift := uintptr(0)
	for i := uintptr(0); i < size; i++ {
		loadedValue |= uintptr(*(*byte)(ptr)) << shift
		shift += 8
		ptr = unsafe.Add(ptr, 1)
	}
	return loadedValue
}

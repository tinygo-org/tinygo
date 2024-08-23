//go:build !mips

package reflect

import "unsafe"

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

// storeValue is the inverse of loadValue. It stores a value to a pointer that
// doesn't need to be aligned.
func storeValue(ptr unsafe.Pointer, size, value uintptr) {
	for i := uintptr(0); i < size; i++ {
		*(*byte)(ptr) = byte(value)
		ptr = unsafe.Add(ptr, 1)
		value >>= 8
	}
}

// maskAndShift cuts out a part of a uintptr. Note that the offset may not be 0.
func maskAndShift(value, offset, size uintptr) uintptr {
	mask := ^uintptr(0) >> ((unsafe.Sizeof(uintptr(0)) - size) * 8)
	return (uintptr(value) >> (offset * 8)) & mask
}

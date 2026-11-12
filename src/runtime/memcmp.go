package runtime

import "unsafe"

func memCompareNop(x, y unsafe.Pointer, size uintptr) bool {
	return true
}

//go:nobounds
func memCompareBytes(x, y unsafe.Pointer, size uintptr) bool {
	for size > 0 {
		size--
		if *(*byte)(unsafe.Add(x, size)) != *(*byte)(unsafe.Add(y, size)) {
			return false
		}
	}
	return true
}

//go:nobounds
func memCompareFloat32(x, y *float32, size uintptr) bool {
	return *x == *y
}

//go:nobounds
func memCompareFloat64(x, y *float64, size uintptr) bool {
	return *x == *y
}

//go:nobounds
func memCompareString(x, y *string, size uintptr) bool {
	return *x == *y
}

//go:nobounds
func memCompareInterface(x, y *interface{}, size uintptr) bool {
	return *x == *y
}

package main

// Test new language features introduced in Go 1.17:
// https://tip.golang.org/doc/go1.17#language
// Once this becomes the minimum Go version of TinyGo, these tests should be
// merged with the regular slice tests.

import "unsafe"

func main() {
	// Test conversion from array to slice.
	slice1 := []int{1, 2, 3, 4}
	arr1 := (*[4]int)(slice1)
	arr1[1] = -2
	arr1[2] = 20
	println("slice to array pointer:", arr1[0], arr1[1], arr1[2], arr1[3])

	// Test unsafe.Add.
	arr2 := [...]int{1, 2, 3, 4}
	*(*int)(unsafe.Add(unsafe.Pointer(&arr2[0]), unsafe.Sizeof(int(1))*1)) = 5
	*addInt(&arr2[0], 2) = 8
	println("unsafe.Add array:", arr2[0], arr2[1], arr2[2], arr2[3])

	// Test unsafe.Slice.
	arr3 := [...]int{1, 2, 3, 4}
	slice3 := unsafe.Slice(&arr3[1], 3)
	slice3[0] = 9
	slice3[1] = 15
	println("unsafe.Slice array:", len(slice3), cap(slice3), slice3[0], slice3[1], slice3[2])
}

func addInt(ptr *int, index uintptr) *int {
	return (*int)(unsafe.Add(unsafe.Pointer(ptr), unsafe.Sizeof(int(1))*index))
}

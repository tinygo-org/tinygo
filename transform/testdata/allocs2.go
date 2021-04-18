package main

func main() {
	n1 := 5
	derefInt(&n1)

	// This should eventually be modified to not escape.
	n2 := 6 // OUT: object allocated on the heap: escapes at line 9
	returnIntPtr(&n2)

	s1 := make([]int, 3)
	readIntSlice(s1)

	s2 := [3]int{}
	readIntSlice(s2[:])

	// This should also be modified to not escape.
	s3 := make([]int, 3) // OUT: object allocated on the heap: escapes at line 19
	returnIntSlice(s3)

	_ = make([]int, getUnknownNumber()) // OUT: object allocated on the heap: size is not constant

	s4 := make([]byte, 300) // OUT: object allocated on the heap: object size 300 exceeds maximum stack allocation size 256
	readByteSlice(s4)
}

func derefInt(x *int) int {
	return *x
}

func returnIntPtr(x *int) *int {
	return x
}

func readIntSlice(s []int) int {
	return s[1]
}

func readByteSlice(s []byte) byte {
	return s[1]
}

func returnIntSlice(s []int) []int {
	return s
}

func getUnknownNumber() int

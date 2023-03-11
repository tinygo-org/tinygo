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

	useSlice(make([]int, getUnknownNumber())) // OUT: object allocated on the heap: size is not constant

	s4 := make([]byte, 300) // OUT: object allocated on the heap: object size 300 exceeds maximum stack allocation size 256
	readByteSlice(s4)

	s5 := make([]int, 4) // OUT: object allocated on the heap: escapes at line 27
	_ = append(s5, 5)

	s6 := make([]int, 3)
	s7 := []int{1, 2, 3}
	copySlice(s6, s7)

	c1 := getComplex128() // OUT: object allocated on the heap: escapes at line 34
	useInterface(c1)

	n3 := 5
	func() int {
		return n3
	}()

	callVariadic(3, 5, 8) // OUT: object allocated on the heap: escapes at line 41

	s8 := []int{3, 5, 8} // OUT: object allocated on the heap: escapes at line 44
	callVariadic(s8...)

	n4 := 3 // OUT: object allocated on the heap: escapes at line 48
	n5 := 7 // OUT: object allocated on the heap: escapes at line 48
	func() {
		n4 = n5
	}()
	println(n4, n5)
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

func copySlice(out, in []int) {
	copy(out, in)
}

func getComplex128() complex128

func useInterface(interface{})

func callVariadic(...int)

func useSlice([]int)

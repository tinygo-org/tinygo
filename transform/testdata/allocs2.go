package main

import (
	"runtime/volatile"
	"unsafe"
)

func main() {
	n1 := 5
	derefInt(&n1)

	// This should eventually be modified to not escape.
	n2 := 6
	returnIntPtr(&n2)

	s1 := make([]int, 3)
	readIntSlice(s1)

	s2 := [3]int{}
	readIntSlice(s2[:])

	// This should also be modified to not escape.
	s3 := make([]int, 3)
	returnIntSlice(s3)

	useSlice(make([]int, getUnknownNumber()))

	s4 := make([]byte, 300)
	readByteSlice(s4)

	s5 := make([]int, 4)
	_ = append(s5, 5)

	s6 := make([]int, 3)
	s7 := []int{1, 2, 3}
	copySlice(s6, s7)

	c1 := getComplex128()
	useInterface(c1)

	n3 := 5
	func() int {
		return n3
	}()

	callVariadic(3, 5, 8)

	s8 := []int{3, 5, 8}
	callVariadic(s8...)

	n4 := 3
	n5 := 7
	func() {
		n4 = n5
	}()
	println(n4, n5)

	// This shouldn't escape.
	var buf [32]byte
	s := string(buf[:])
	println(len(s))

	var rbuf [5]rune
	s = string(rbuf[:])
	println(s)

	// Unsafe usage of DMA buffers: the compiler thinks this buffer won't be
	// used anymore after the volatile store.
	var dmaBuf1 [4]byte
	pseudoVolatile.Set(uint32(unsafeNoEscape(unsafe.Pointer(&dmaBuf1[0]))))

	// Safe usage of DMA buffers: keep the buffer alive until it is no longer
	// needed, but don't mark it as needing to be heap allocated. The compiler
	// will keep the buffer stack allocated if possible.
	var dmaBuf2 [4]byte
	pseudoVolatile.Set(uint32(unsafeNoEscape(unsafe.Pointer(&dmaBuf2[0]))))
	// ...use the buffer in the DMA peripheral
	keepAliveNoEscape(unsafe.Pointer(&dmaBuf2[0]))
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

// See the function with the same name in the machine package.
//
//go:linkname unsafeNoEscape machine.unsafeNoEscape
func unsafeNoEscape(ptr unsafe.Pointer) uintptr

//go:linkname keepAliveNoEscape machine.keepAliveNoEscape
func keepAliveNoEscape(ptr unsafe.Pointer)

var pseudoVolatile volatile.Register32

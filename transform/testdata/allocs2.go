package main

import (
	"fmt"
	"reflect"
	"runtime/volatile"
	"unsafe"
)

func main() {
	n1 := 5
	derefInt(&n1)

	n2 := 6
	returnIntPtr(&n2)

	s1 := make([]int, 3)
	readIntSlice(s1)

	s2 := [3]int{}
	readIntSlice(s2[:])

	s3 := make([]int, 3)
	returnIntSlice(s3)

	useSlice(make([]int, getUnknownNumber())) // OUT: makeslice.buf size is not constant

	s4 := make([]byte, 300) // OUT: makeslice4 size 300 exceeds maximum stack allocation size 256
	readByteSlice(s4)

	s5 := make([]int, 4) // OUT: makeslice6 escapes at line 32
	_ = append(s5, 5)

	s6 := make([]int, 3)
	s7 := []int{1, 2, 3}
	copySlice(s6, s7)

	c1 := getComplex128() // OUT: Arg 0 (type: basic:complex128) of main.useInterface() escapes at line 40

	useInterface(c1)

	a1 := [3]int32{10, 11, 12}
	useInterface(a1) // OUT: Arg 0 (type: array:3:basic:int32) of main.useInterface() escapes at line 43

	n3 := 5
	func() int {
		return n3
	}()

	callVariadic(3, 5, 8) // OUT: varargs12 escapes at line 50

	s8 := []int{3, 5, 8} // OUT: slicelit14 escapes at line 53
	callVariadic(s8...)

	n4 := 3 // OUT: n4 escapes at line 57
	n5 := 7 // OUT: n5 escapes at line 57
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
	// Doesn't escape.
	theErr := theError{}
	var err error = theErr
	theErr.set("hello")
	println(err.Error())
}

type vector3 [3]float32

func scaleVector3(vec *vector3, f float32) *vector3 {
	vec[0] *= f
	vec[1] *= f
	vec[2] *= f
	return vec
}

func crossVector3(a, b *vector3) vector3 {
	return vector3{
		a[1]*b[2] - a[2]*b[1],
		a[2]*b[0] - a[0]*b[2],
		a[0]*b[1] - a[1]*b[0],
	}
}

type msgSetter interface {
	set(msg string)
}

type theError struct {
	msg string
}

func makeTheError(msg string) theError { return theError{msg: msg} }

func (this theError) Error() string { return this.msg }

func (this theError) set(msg string) { this.msg = msg }

func nonEscapingReturnedPointer() vector3 {
	a := vector3{1, 2, 3}
	b := vector3{4, 5, 6}

	c := scaleVector3(&b, 0.5)
	return crossVector3(&a, c)
}

var escapedSlice []int

func escapingReturnedSlice() {
	s := make([]int, 3) // OUT: makeslice escapes at line 133
	escapedSlice = returnIntSlice(s)
}

var escapedVector3 *vector3

func escapingReturnedPointer() {
	b := vector3{4, 5, 6} // OUT: b escapes at line 142

	c := scaleVector3(&b, 0.5)
	escapedVector3 = c
}

var escapedArray [2](*vector3)

func escapingCompositeLiterals() {
	c := scaleVector3(&vector3{4, 5, 6}, 0.5) // OUT: Arg 0 of main.scaleVector3() escapes at line 149
	escapedVector3 = c
	a := [2](*vector3){&vector3{7, 8, 9}, &vector3{2, 3, 4}} // OUT: complit1 escapes at line 151 && complit2 escapes at line 151
	escapedArray = a
}

func escapingInterfaceArguments() {
	// Check that the argument indices in the escape messages are right.
	useInterfaces(16, &vector3{7, 8, 9}, reflect.TypeFor[int](), 32, &vector3{6, 7, 8}) // OUT: Arg 1 (type: pointer:named:main.vector3) of main.useInterfaces() escapes at line 156 && Arg 4 (type: pointer:named:main.vector3) of main.useInterfaces() escapes at line 156
}

func passAnInterface(t reflect.Type) {
	useInterfaces(16, &vector3{7, 8, 9}, t, 32, &vector3{6, 7, 8}) // OUT: Arg 1 (type: pointer:named:main.vector3) of main.useInterfaces() escapes at line 160 && Arg 4 (type: pointer:named:main.vector3) of main.useInterfaces() escapes at line 160
}

func recursiveScaleVector3(vec *vector3, n int) *vector3 {
	if n == 0 {
		return vec
	}
	return recursiveScaleVector3(vec, n-1)
}

func recursiveReturnedPointer() vector3 {
	b := vector3{4, 5, 6} // OUT: b escapes at unknown line

	c := recursiveScaleVector3(&b, 1)
	return *c
}

func theFmt() string {
	v := 48
	res := fmt.Sprintf("The number is %d", v) // OUT: varargs escapes at line 179
	return res
}

func giveAnError() error {
	return theError{msg: "hello"} // OUT: main.theError escapes at line 184
}

func giveOtherError() error {
	return makeTheError("wellcome") // OUT: main.theError escapes at line 188
}

func giveAnInterface() interface{} {
	a := [3]int32{1, 2, 3}
	return a // OUT: array:3:basic:int32 escapes at line 193
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

func useInterfaces(int, interface{}, reflect.Type, int, interface{})

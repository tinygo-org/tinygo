package main

import (
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

	useSlice(make([]int, getUnknownNumber())) // OUT: size is not constant

	s4 := make([]byte, 300) // OUT: object size 300 exceeds maximum stack allocation size 256
	readByteSlice(s4)

	s5 := make([]int, 4) // OUT: escapes at line 30
	_ = append(s5, 5)

	s6 := make([]int, 3)
	s7 := []int{1, 2, 3}
	copySlice(s6, s7)

	c1 := getComplex128() // OUT: escapes at line 37
	useInterface(c1)

	n3 := 5
	func() int {
		return n3
	}()

	callVariadic(3, 5, 8) // OUT: escapes at line 44

	s8 := []int{3, 5, 8} // OUT: escapes at line 47
	callVariadic(s8...)

	n4 := 3 // OUT: escapes at line 51
	n5 := 7 // OUT: escapes at line 51
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

func nonEscapingReturnedPointer() vector3 {
	a := vector3{1, 2, 3}
	b := vector3{4, 5, 6}

	c := scaleVector3(&b, 0.5)
	return crossVector3(&a, c)
}

var escapedSlice []int

func escapingReturnedSlice() {
	s := make([]int, 3) // OUT: escapes at line 108
	escapedSlice = returnIntSlice(s)
}

var escapedVector3 *vector3

func escapingReturnedPointer() {
	b := vector3{4, 5, 6} // OUT: escapes at line 117

	c := scaleVector3(&b, 0.5)
	escapedVector3 = c
}

func recursiveScaleVector3(vec *vector3, n int) *vector3 {
	if n == 0 {
		return vec
	}
	return recursiveScaleVector3(vec, n-1)
}

func recursiveReturnedPointer() vector3 {
	b := vector3{4, 5, 6} // OUT: escapes at unknown line

	c := recursiveScaleVector3(&b, 1)
	return *c
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

type errT struct{ x [4]int }

func (e *errT) Error() string { return "errT" }

type ptrStruct struct {
	n int
	p *errT
}

func wrapError(e *errT) (int, error) { return 1, e }

func wrapErrorFirst(e *errT) (error, int) { return e, 1 }

func wrapAny(e *errT) (int, any) { return 1, e }

func wrapSlice(b []byte) (int, []byte) { return 1, b }

func wrapStruct(e *errT) (int, ptrStruct) { return 1, ptrStruct{1, e} }

func wrapArray(e *errT) (int, [2]*errT) { return 1, [2]*errT{e, nil} }

var globalErr error

// The pointer is returned as one of multiple return values, wrapped in an
// interface, slice, struct or array, and then escapes from the caller.
func escapingMultiReturn() error {
	e := &errT{} // OUT: escapes at line 206
	_, err := wrapError(e)
	return err
}

func escapingMultiReturnFirst() error {
	e := &errT{} // OUT: escapes at line 212
	err, _ := wrapErrorFirst(e)
	return err
}

func escapingMultiReturnAny() any {
	e := &errT{} // OUT: escapes at line 218
	_, v := wrapAny(e)
	return v
}

func escapingMultiReturnSlice() []byte {
	b := make([]byte, 8) // OUT: escapes at line 224
	_, s := wrapSlice(b)
	return s
}

func escapingMultiReturnStruct() *errT {
	e := &errT{} // OUT: escapes at line 195
	_, s := wrapStruct(e)
	return s.p
}

func escapingMultiReturnArray() *errT {
	e := &errT{} // OUT: escapes at line 236
	_, a := wrapArray(e)
	return a[0]
}

func escapingMultiReturnGlobal() {
	e := &errT{} // OUT: escapes at line 242
	_, err := wrapError(e)
	globalErr = err
}

// The pointer is returned as one of multiple return values, but the caller
// does not let it escape, so it can stay on the stack.
func nonEscapingMultiReturnInt() int {
	e := &errT{}
	n, _ := wrapError(e)
	return n
}

func nonEscapingMultiReturnNilCheck() bool {
	e := &errT{}
	_, err := wrapError(e)
	return err != nil
}

func nonEscapingMultiReturnSliceLen() int {
	b := make([]byte, 8)
	_, s := wrapSlice(b)
	return len(s)
}

func nonEscapingMultiReturnArrayLoad() int {
	e := &errT{}
	e.x[1] = 42
	_, a := wrapArray(e)
	return a[0].x[1]
}

func wrapSingleArray(e *errT) [1]error { return [1]error{e} }

func wrapNestedArray(e *errT) (int, [2][2]*errT) { return 1, [2][2]*errT{{}, {nil, e}} }

func forwardError(e *errT) error {
	_, err := wrapError(e)
	return err
}

func escapingSingleArray() error {
	e := &errT{} // OUT: escapes at line 284
	a := wrapSingleArray(e)
	return a[0]
}

func escapingNestedArray() [2]*errT {
	e := &errT{} // OUT: escapes at line 290
	_, a := wrapNestedArray(e)
	return a[1]
}

func escapingForwardedError() error {
	e := &errT{} // OUT: escapes at line 295
	return forwardError(e)
}

func escapingMultiReturnField() *int {
	e := &errT{} // OUT: escapes at line 301
	_, a := wrapArray(e)
	return &a[0].x[1]
}

func escapingMultiReturnCall() {
	e := &errT{} // OUT: escapes at line 307
	_, err := wrapError(e)
	useInterface(err)
}

func escapingMultiReturnClosure() func() error {
	e := &errT{}           // OUT: escapes at line 312
	_, err := wrapError(e) // OUT: escapes at line 313
	return func() error { return err }
}

func escapingMultiReturnPhi(cond bool) error {
	e := &errT{} // OUT: escapes at line 318
	_, err := wrapError(e)
	if cond {
		err = nil
	}
	return err
}

func nonEscapingSingleArray() bool {
	e := &errT{}
	a := wrapSingleArray(e)
	return a[0] != nil
}

func nonEscapingNestedArray() int {
	e := &errT{}
	_, a := wrapNestedArray(e)
	return a[1][1].x[0]
}

func nonEscapingForwardedError() bool {
	e := &errT{}
	return forwardError(e) != nil
}

func nonEscapingMultiReturnSliceCap() int {
	b := make([]byte, 8)
	_, s := wrapSlice(b)
	return cap(s)
}

func nonEscapingMultiReturnDiscard() {
	e := &errT{}
	wrapError(e)
}

func wrapPointerFreeArray(e *errT) (error, [2]int) { return e, [2]int{e.x[0], 1} }

func nonEscapingPointerFreeArray() [2]int {
	e := &errT{}
	_, a := wrapPointerFreeArray(e)
	return a
}

package main

// Basic tests that don't need to be split into a separate file.

func addInt(x, y int) int {
	return x + y
}

func equalInt(x, y int) bool {
	return x == y
}

func equalByteArray(x, y *[4096]byte) bool {
	return *x == *y
}

func notEqualByteArray(x, y *[4096]byte) bool {
	return *x != *y
}

func equalMutatedByteArray(x *[5]byte) bool {
	value := *x
	x[0] = 1
	return value == *x
}

func divInt(x, y int) int {
	return x / y
}

func divUint(x, y uint) uint {
	return x / y
}

func remInt(x, y int) int {
	return x % y
}

func remUint(x, y uint) uint {
	return x % y
}

func floatEQ(x, y float32) bool {
	return x == y
}

func floatNE(x, y float32) bool {
	return x != y
}

func floatLower(x, y float32) bool {
	return x < y
}

func floatLowerEqual(x, y float32) bool {
	return x <= y
}

func floatGreater(x, y float32) bool {
	return x > y
}

func floatGreaterEqual(x, y float32) bool {
	return x >= y
}

func complexReal(x complex64) float32 {
	return real(x)
}

func complexImag(x complex64) float32 {
	return imag(x)
}

func complexAdd(x, y complex64) complex64 {
	return x + y
}

func complexSub(x, y complex64) complex64 {
	return x - y
}

func shiftNested(x uint64) uint64 {
	k := 3
	return x >> (1 << k) // https://github.com/tinygo-org/tinygo/issues/5496
}

func complexMul(x, y complex64) complex64 {
	return x * y
}

// TODO: complexDiv (requires runtime call)

// A type 'kv' also exists in function foo. Test that these two types don't
// conflict with each other.
type kv struct {
	v       float32
	x, y, z int
}

var kvGlobal kv

func foo() {
	// Define a new 'kv' type.
	type kv struct {
		v       byte
		x, y, z int
	}
	// Use this type.
	func(b kv) {}(kv{})
}

type T1 []T1
type T2 [2]*T2

var a T1
var b T2

// Zero-length array fields do not count towards the fields of an expanded
// parameter.
type zeroArrayFields struct {
	a [0]int
	b int32
	c [0]byte
	d int32
	e int32
	f [0]struct{}
}

type zeroArrayTooMany struct {
	a [0]int
	b int32
	c int32
	d int32
	e int32
}

func zeroArrayParam(s zeroArrayFields) int32 {
	return s.b + s.d + s.e
}

func zeroArrayParamTooMany(s zeroArrayTooMany) int32 {
	return s.b + s.e
}

func callZeroArrayParams() int32 {
	return zeroArrayParam(zeroArrayFields{b: 1, d: 2, e: 3}) + zeroArrayParamTooMany(zeroArrayTooMany{b: 4, e: 5})
}

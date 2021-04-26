package main

// This file tests the type codes assigned by the reflect lowering pass.
// This test is not complete, most importantly, sidetables are not currently
// being tested.

import (
	"reflect"
	"unsafe"
)

const (
	// See the top of src/reflect/type.go
	prefixChan      = 0b0001
	prefixInterface = 0b0011
	prefixPtr       = 0b0101
	prefixSlice     = 0b0111
	prefixArray     = 0b1001
	prefixFunc      = 0b1011
	prefixMap       = 0b1101
	prefixStruct    = 0b1111
)

func main() {
	// Check for some basic types.
	assertType(3, uintptr(reflect.Int)<<1)
	assertType(uint8(3), uintptr(reflect.Uint8)<<1)
	assertType(byte(3), uintptr(reflect.Uint8)<<1)
	assertType(int64(3), uintptr(reflect.Int64)<<1)
	assertType("", uintptr(reflect.String)<<1)
	assertType(3.5, uintptr(reflect.Float64)<<1)
	assertType(unsafe.Pointer(nil), uintptr(reflect.UnsafePointer)<<1)

	// Check for named types: they are given names in order.
	// They are sorted in reverse, for no good reason.
	const intNum = uintptr(reflect.Int) << 1
	assertType(namedInt1(0), (3<<6)|intNum)
	assertType(namedInt2(0), (2<<6)|intNum)
	assertType(namedInt3(0), (1<<6)|intNum)

	// Check for some "prefix-style" types.
	assertType(make(chan int), (intNum<<5)|prefixChan)
	assertType(new(int), (intNum<<5)|prefixPtr)
	assertType([]int{}, (intNum<<5)|prefixSlice)
}

type (
	namedInt1 int
	namedInt2 int
	namedInt3 int
)

// Pseudo call that is being checked by the code in reflect_test.go.
// After reflect lowering, the type code as part of the interface should match
// the asserted type code.
func assertType(itf interface{}, assertedTypeCode uintptr)

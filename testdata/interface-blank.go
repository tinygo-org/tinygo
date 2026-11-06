package main

import "unsafe"

type blankStruct struct {
	_     int `first:"blank"`
	Value int
	_     int `last:"blank"`
}

type nestedStruct struct {
	Value blankStruct
}

type largeBlankStruct struct {
	_     [130]byte
	Value byte
	_     byte
}

func equal(x, y interface{}) bool {
	return x == y
}

func main() {
	x := *(*blankStruct)(unsafe.Pointer(&[3]int{1, 2, 3}))
	y := *(*blankStruct)(unsafe.Pointer(&[3]int{4, 2, 5}))
	z := *(*blankStruct)(unsafe.Pointer(&[3]int{1, 6, 3}))

	println("blank struct equality:", equal(x, y), equal(x, z))
	println("blank array equality:", equal([1]blankStruct{x}, [1]blankStruct{y}), equal([1]blankStruct{x}, [1]blankStruct{z}))
	println("nested blank struct equality:", equal(nestedStruct{x}, nestedStruct{y}), equal(nestedStruct{x}, nestedStruct{z}))

	largeX := *(*largeBlankStruct)(unsafe.Pointer(&[132]byte{0: 1, 130: 2, 131: 3}))
	largeY := *(*largeBlankStruct)(unsafe.Pointer(&[132]byte{0: 4, 130: 2, 131: 5}))
	largeZ := *(*largeBlankStruct)(unsafe.Pointer(&[132]byte{0: 1, 130: 6, 131: 3}))
	println("large blank struct equality:", equal(largeX, largeY), equal(largeX, largeZ))
}

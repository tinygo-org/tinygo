package main

import "unsafe"

func sliceLen(ints []int) int {
	return len(ints)
}

func sliceCap(ints []int) int {
	return cap(ints)
}

func sliceElement(ints []int, index int) int {
	return ints[index]
}

func sliceAppendValues(ints []int) []int {
	return append(ints, 1, 2, 3)
}

func sliceAppendSlice(ints, added []int) []int {
	return append(ints, added...)
}

func sliceCopy(dst, src []int) int {
	return copy(dst, src)
}

// Test bounds checking in *ssa.MakeSlice instruction.

func makeByteSlice(len int) []byte {
	return make([]byte, len)
}

func makeInt16Slice(len int) []int16 {
	return make([]int16, len)
}

func makeArraySlice(len int) [][3]byte {
	return make([][3]byte, len) // slice with element size of 3
}

func makeInt32Slice(len int) []int32 {
	return make([]int32, len)
}

func Add32(p unsafe.Pointer, len int) unsafe.Pointer {
	return unsafe.Add(p, len)
}

func Add64(p unsafe.Pointer, len int64) unsafe.Pointer {
	return unsafe.Add(p, len)
}

func SliceToArray(s []int) *[4]int {
	return (*[4]int)(s)
}

func SliceToArrayConst() *[4]int {
	s := make([]int, 6)
	return (*[4]int)(s)
}

func SliceInt(ptr *int, len int) []int {
	return unsafe.Slice(ptr, len)
}

func SliceUint16(ptr *byte, len uint16) []byte {
	return unsafe.Slice(ptr, len)
}

func SliceUint64(ptr *int, len uint64) []int {
	return unsafe.Slice(ptr, len)
}

func SliceInt64(ptr *int, len int64) []int {
	return unsafe.Slice(ptr, len)
}

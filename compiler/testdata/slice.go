package main

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

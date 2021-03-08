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

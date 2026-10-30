package main

// A slice or string without capacity keeps the base pointer, so that it never
// points at the next object in memory.

func sliceTail(ints []int, low int) []int {
	return ints[low:]
}

func sliceTailMax(ints []int, low, high, max int) []int {
	return ints[low:high:max]
}

func sliceArrayTail(a *[4]int, low int) []int {
	return a[low:]
}

func stringTail(s string, low int) string {
	return s[low:]
}

// The capacity is known not to be zero, so no select is needed.
func sliceConstMax(ints []int, low int) []int {
	return ints[low : low+1 : low+2]
}

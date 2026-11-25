package main

// Range loops over slices compare the index with an unsigned compare. The
// index and the length are never negative, so it gives the same result.

func rangeCount(s []int) int {
	n := 0
	for range s {
		n++
	}
	return n
}

func rangeIndexSum(s []byte) int {
	n := 0
	for i := range s {
		n += i
	}
	return n
}

// Other loops keep the signed compare of the Go source.

func countUp(s []int) int {
	n := 0
	for i := 0; i < len(s); i++ {
		n += i
	}
	return n
}

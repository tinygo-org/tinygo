package main

import "os"

//go:nopanic index
func proven(s []int) int {
	sum := 0
	for i := 0; i < len(s); i++ {
		sum += s[i]
	}
	return sum
}

//go:nopanic index
func unproven(s []int, i int) int {
	return s[i] + s[i+1]
}

func main() {
	s := []int{1, 2, 3}
	println(proven(s), unproven(s, len(os.Args)), unproven(s, len(os.Args)-1))
}

// ERROR: nopanic-check.go:16:10: //go:nopanic check is not proven
// ERROR: nopanic-check.go:16:17: //go:nopanic check is not proven

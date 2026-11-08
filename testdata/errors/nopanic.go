package main

//go:nopanic
func noKinds() {}

//go:nopanic index slice
func twoParams() {}

//go:nopanic index,bounds
func unknownKind() {}

//go:nopanic index
func declaration()

//go:nopanic index
func valid(s []int) int {
	sum := 0
	for i := 0; i < len(s); i++ {
		sum += s[i]
	}
	return sum
}

func main() {
	noKinds()
	twoParams()
	unknownKind()
	declaration()
	println(valid([]int{1, 2, 3}))
}

// ERROR: # command-line-arguments
// ERROR: nopanic.go:3:1: expected one parameter to //go:nopanic, not 0
// ERROR: nopanic.go:6:1: expected one parameter to //go:nopanic, not 2
// ERROR: nopanic.go:9:1: unknown //go:nopanic kind "bounds"
// ERROR: nopanic.go:12:1: can only use //go:nopanic on definitions

package main

type hasPadding struct {
	b1 bool
	i  int
	b2 bool
}

type nestedPadding struct {
	b bool
	hasPadding
	i int
}

//go:noinline
func testZeroGet(m map[hasPadding]int, s hasPadding) int {
	return m[s]
}

//go:noinline
func testZeroSet(m map[hasPadding]int, s hasPadding) {
	m[s] = 5
}

//go:noinline
func testZeroArrayGet(m map[[2]hasPadding]int, s [2]hasPadding) int {
	return m[s]
}

//go:noinline
func testZeroArraySet(m map[[2]hasPadding]int, s [2]hasPadding) {
	m[s] = 5
}

func main() {

}

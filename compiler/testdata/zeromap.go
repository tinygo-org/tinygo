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

type stringStruct struct {
	a string
	b string
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

//go:noinline
func makeStringStructMap() map[stringStruct]int {
	return make(map[stringStruct]int)
}

//go:noinline
func makeShortStringArrayMap() map[[2]string]int {
	return make(map[[2]string]int)
}

//go:noinline
func makeLongStringArrayMap() map[[5]string]int {
	return make(map[[5]string]int)
}

func main() {

}

package main

//go:noheap
func main() {
	// This object is optimized away, and won't cause a linker failure.
	var a int
	add(&a)

	// This object is not optimized away since its address is captured.
	var b int
	escape(&b)

	// A simple defer won't cause a heap allocation.
	defer func() {
		println("once")
	}()

	// A defer in a loop does causes a heap allocation.
	for range 3 {
		defer func() {
			println("many")
		}()
	}

	// This slice should not be heap-allocated.
	s1 := make([]int, 5)
	sum(s1)

	// But this slice is.
	s2 := make([]int, 5)
	escape(&s2[0])

	// Allocated at compile time as a global, so won't escape at runtime.
	escape(globalInt)
}

func add(n *int) {
	*n++
}

func sum(slice []int) (result int) {
	for n := range slice {
		result += n
	}
	return result
}

var globalInt = globalNewInt()

var globalAssign *int

//go:noheap
func globalNewInt() *int {
	return new(int)
}

func escape(n *int) {
	// Do some stuff that will definitely force a heap allocation.
	// (This might be optimized in the future, in which case we need to change
	// the function again).
	n2 := func(n *int) *int {
		return n
	}(n)
	println(n2)
}

// ERROR: noheap.go:10: object allocated on the heap in //go:noheap function
// ERROR: noheap.go:20: object allocated on the heap in //go:noheap function
// ERROR: noheap.go:30: object allocated on the heap in //go:noheap function

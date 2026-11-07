package main

// Hosted targets grow slices as nextslicecap in Go's runtime/slice.go does,
// without rounding to a size class. Bare metal and the leaking GC double.
func main() {
	var s []int
	last := -1
	for i := 0; i < 5000; i++ {
		s = append(s, i)
		if cap(s) != last {
			last = cap(s)
			println("len", len(s), "cap", last)
		}
	}
	s = append(s[:10:10], make([]int, 100)...)
	println("append many: len", len(s), "cap", cap(s))
}

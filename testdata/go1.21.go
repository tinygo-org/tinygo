package main

func main() {
	// The new min/max builtins.
	ia := 1
	ib := 5
	ic := -3
	fa := 1.0
	fb := 5.0
	fc := -3.0
	println("min/max:", min(ia, ib, ic), max(ia, ib, ic))
	println("min/max:", min(fa, fb, fc), max(fa, fb, fc))

	// The clear builtin, for slices.
	s := []int{1, 2, 3, 4, 5}
	clear(s[:3])
	println("cleared s[:3]:", s[0], s[1], s[2], s[3], s[4])
}

package main

import "math"

func main() {
	// The new min/max builtins.
	// With int:
	ia := 1
	ib := 5
	ic := -3
	println("min/max:", min(ia, ib, ic), max(ia, ib, ic))
	// With float:
	fa := 1.0
	fb := 5.0
	fc := -3.0
	println("min/max:", min(fa, fb, fc), max(fa, fb, fc))
	// Float +/- 0.0:
	pos0 := 0.0
	neg0 := -pos0
	println("min/max:", min(pos0, neg0), max(pos0, neg0))
	// Float NaN:
	println("min/max:", min(math.NaN(), 12.0), max(math.NaN(), 12.0))

	// The clear builtin, for slices.
	s := []int{1, 2, 3, 4, 5}
	clear(s[:3])
	println("cleared s[:3]:", s[0], s[1], s[2], s[3], s[4])

	// The clear builtin, for maps.
	m := map[int]string{
		1: "one",
		2: "two",
		3: "three",
	}
	clear(m)
	println("cleared map:", m[1], m[2], m[3], len(m))
	m[4] = "four"
	println("added to cleared map:", m[1], m[2], m[3], m[4], len(m))
}

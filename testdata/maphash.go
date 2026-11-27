package main

import (
	"hash/maphash"
	"math"
)

type point struct {
	x, y int
	name string
}

type shape interface {
	area() float64
}

type square struct{ side float64 }

func (s square) area() float64 { return s.side * s.side }

func main() {
	seed := maphash.MakeSeed()

	// Equal values must have equal hashes.
	s1 := "hello"
	s2 := string([]byte("hello"))
	println("string:", maphash.Comparable(seed, s1) == maphash.Comparable(seed, s2))
	println("int:", maphash.Comparable(seed, 42) == maphash.Comparable(seed, 42))
	println("struct:", maphash.Comparable(seed, point{1, 2, s1}) == maphash.Comparable(seed, point{1, 2, s2}))
	println("array:", maphash.Comparable(seed, [3]int{1, 2, 3}) == maphash.Comparable(seed, [3]int{1, 2, 3}))
	zero, negZero := 0.0, math.Copysign(0, -1)
	println("float zero:", maphash.Comparable(seed, zero) == maphash.Comparable(seed, negZero))
	var a, b shape = square{2}, square{2}
	println("interface:", maphash.Comparable(seed, a) == maphash.Comparable(seed, b))
	v := 7
	p := &v
	println("pointer:", maphash.Comparable(seed, p) == maphash.Comparable(seed, &v))

	// Different values should have different hashes.
	println("different string:", maphash.Comparable(seed, "a") != maphash.Comparable(seed, "b"))
	println("different struct:", maphash.Comparable(seed, point{1, 2, "a"}) != maphash.Comparable(seed, point{2, 1, "a"}))

	// The seed must change the hash.
	println("different seed:", maphash.Comparable(seed, s1) != maphash.Comparable(maphash.MakeSeed(), s1))

	// WriteComparable must agree for equal values.
	var h1, h2 maphash.Hash
	h1.SetSeed(seed)
	h2.SetSeed(seed)
	maphash.WriteComparable(&h1, point{3, 4, s1})
	maphash.WriteComparable(&h2, point{3, 4, s2})
	println("WriteComparable:", h1.Sum64() == h2.Sum64())
}

package main

type empty struct{}

type pair struct{ a, b uint64 }

type nested struct{ e empty }

type fns struct{}

// v1 has a debug value and is used again by the return.
func (fns) merge(_ pair, v1 empty, _ pair, v2 empty) (empty, bool) {
	return v1, v1 == v2
}

func zeroArray(a, b [0]uint64) ([0]uint64, bool) {
	return a, a == b
}

func nestedEmpty(n nested) (nested, int) {
	m := n
	return m, 7
}

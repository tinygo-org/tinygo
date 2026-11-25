package main

import "runtime"

type T struct{ a, b, c, d uint16 }

var sink *T

//go:noinline
func newT(v uint16) *T {
	t := new(T)
	t.a, t.b, t.c, t.d = v, v, v, v
	return t
}

// check holds p only in a register while the collector runs, then reuses freed
// memory. The block of p is overwritten if the collector missed it.
//
//go:noinline
func check(v uint16) bool {
	p := newT(v)
	runtime.GC()
	for i := 0; i < 200; i++ {
		sink = newT(0xAAAA)
	}
	return p.a == v && p.b == v && p.c == v && p.d == v
}

func main() {
	bad := 0
	for i := 0; i < 20; i++ {
		if !check(uint16(0x1000 + i)) {
			bad++
		}
	}
	if bad != 0 {
		println("wrong: live object was freed and reused", bad, "of 20")
	} else {
		println("ok")
	}
}

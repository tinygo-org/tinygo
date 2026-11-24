package main

// Objects held only through a slice or string of zero capacity at their end
// must stay alive, and that slice must not point into the next object.

import (
	"runtime"
	"unsafe"
)

type blob [4096]byte

type record struct {
	i int
	b []byte
}

var (
	tailFinalized    int
	droppedFinalized int
	sink             int
	held             []record
	heldTails        [][]byte
)

//go:noinline
func index(n int) int { return n }

//go:noinline
func scrubStack(depth int) int {
	if depth <= 0 {
		return sink
	}
	var buf [64]int
	for i := range buf {
		buf[i] = depth + i
	}
	sink += buf[depth&63]
	return scrubStack(depth-1) + buf[0]
}

// tailOf returns only the end of a new blob that has a finalizer.
//
//go:noinline
func tailOf() []byte {
	p := new(blob)
	p[len(p)-1] = 7
	runtime.SetFinalizer(p, func(*blob) { tailFinalized++ })
	return p[index(len(p)):]
}

//go:noinline
func allocAndDrop() {
	p := new(blob)
	runtime.SetFinalizer(p, func(*blob) { droppedFinalized++ })
}

func checkPointers() {
	b := make([]byte, index(4096))
	base := unsafe.SliceData(b)
	if unsafe.SliceData(b[index(len(b)):]) != base {
		println("slice end does not keep its base")
	}
	if unsafe.SliceData(b[index(4096):index(4096):index(4096)]) != base {
		println("three index slice end does not keep its base")
	}
	if unsafe.SliceData(b[index(1):]) != (*byte)(unsafe.Add(unsafe.Pointer(base), 1)) {
		println("slice with capacity lost its offset")
	}
	if c := b[index(10):index(10)]; cap(c) == 0 || unsafe.SliceData(c) != (*byte)(unsafe.Add(unsafe.Pointer(base), 10)) {
		println("empty slice with capacity lost its offset")
	}
	a := new(blob)
	if unsafe.SliceData(a[index(len(a)):]) != &a[0] {
		println("array end does not keep its base")
	}
	s := string(b[:index(100)])
	if unsafe.StringData(s[index(len(s)):]) != unsafe.StringData(s) {
		println("string end does not keep its base")
	}
	if unsafe.StringData(s[index(3):]) != (*byte)(unsafe.Add(unsafe.Pointer(unsafe.StringData(s)), 3)) {
		println("string lost its offset")
	}
}

func main() {
	checkPointers()

	// Exact power of two objects next to each other, half of them held only
	// through their end.
	for i := 0; i < 64; i++ {
		b := make([]byte, index(4096))
		for j := range b {
			b[j] = byte(i + j)
		}
		if i%2 == 0 {
			held = append(held, record{i, b})
		} else {
			heldTails = append(heldTails, b[index(len(b)):])
		}
		small := make([]byte, index(64))
		for j := range small {
			small[j] = byte(i * j)
		}
		held = append(held, record{i, small})
	}
	tail := tailOf()
	allocAndDrop()
	for i := 0; i < 100 && droppedFinalized == 0; i++ {
		sink += scrubStack(40)
		runtime.GC()
		runtime.Gosched()
		_ = make([]byte, index(4096))
	}
	if droppedFinalized == 0 {
		println("dropped object was never finalized")
	}
	if tailFinalized != 0 {
		println("object held through its end was finalized")
	}
	if len(tail) != 0 || cap(tail) != 0 || unsafe.SliceData(tail) == nil {
		println("bad tail")
	}
	if (*blob)(unsafe.Pointer(unsafe.SliceData(tail)))[4095] != 7 {
		println("object held through its end lost its data")
	}
	for _, r := range held {
		for j := range r.b {
			want := byte(r.i + j)
			if len(r.b) == 64 {
				want = byte(r.i * j)
			}
			if r.b[j] != want {
				println("held object changed", r.i, j)
				return
			}
		}
	}
	for _, t := range heldTails {
		if len(t) != 0 || cap(t) != 0 {
			println("bad tail")
		}
	}
	println("ok")
}

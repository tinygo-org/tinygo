package main

import "unsafe"

// Heap objects that hold a pointer stored as uintptr during package
// initialization. Interp must keep the pointer at the same byte offset.

var x int

const size = unsafe.Sizeof(uintptr(0))

type withPointer struct {
	p *int
	a [2]uintptr
}

var (
	aligned    = newAligned()
	misaligned = newMisaligned()
	one        = newOne()
	self       = newSelf()
	straddle   = newStraddle()
)

func newAligned() *[2]uintptr {
	p := new([2]uintptr)
	p[1] = uintptr(unsafe.Pointer(&x))
	return p
}

func newMisaligned() *[2]uintptr {
	p := new([2]uintptr)
	*(*uintptr)(unsafe.Add(unsafe.Pointer(p), size/2)) = uintptr(unsafe.Pointer(&x))
	return p
}

func newOne() *uintptr {
	p := new(uintptr)
	*p = uintptr(unsafe.Pointer(&x))
	return p
}

func newSelf() *[2]uintptr {
	p := new([2]uintptr)
	p[1] = uintptr(unsafe.Pointer(p))
	return p
}

func newStraddle() *withPointer {
	p := &withPointer{p: &x}
	*(*uintptr)(unsafe.Add(unsafe.Pointer(&p.a), size/2)) = uintptr(unsafe.Pointer(&x))
	return p
}

// holdsX reports whether the object at p has the bytes of &x at each offset,
// and zero bytes in other places.
func holdsX(p unsafe.Pointer, n uintptr, offsets ...uintptr) bool {
	v := uintptr(unsafe.Pointer(&x))
	want := make([]byte, n)
	for _, offset := range offsets {
		copy(want[offset:], unsafe.Slice((*byte)(unsafe.Pointer(&v)), size))
	}
	return string(unsafe.Slice((*byte)(p), n)) == string(want)
}

func main() {
	println("aligned:", holdsX(unsafe.Pointer(aligned), 2*size, size))
	println("misaligned:", holdsX(unsafe.Pointer(misaligned), 2*size, size/2))
	println("one:", holdsX(unsafe.Pointer(one), size, 0))
	println("self:", self[0] == 0 && self[1] == uintptr(unsafe.Pointer(self)))
	println("straddle:", straddle.p == &x && holdsX(unsafe.Pointer(&straddle.a), 2*size, size/2))
}

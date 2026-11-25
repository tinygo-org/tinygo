package main

// Closures whose captured variables may or may not be passed by value. Each
// case checks that the closure sees the same value as with a heap cell.

import (
	"errors"
	"runtime"
)

type N int

func (n *N) Inc() { *n++ }
func (n N) Get() int { return int(n) }

type P struct{ a, b int }

var keep []func() int
var gp *int

//go:noinline
func val(i int) int { return i }

// Recursive closure: the closure is the stored value, so it is created before the store.
func recursive() int {
	var f func(int) int
	f = func(n int) int {
		if n == 0 {
			return 0
		}
		return n + f(n-1)
	}
	return f(10)
}

// Closure created on a branch where the store has not run.
func branch(c bool) int {
	var g func() int
	if c {
		x := val(3)
		g = func() int { return x }
	} else {
		g = func() int { return -1 }
	}
	return g()
}

// Named result read by a deferred closure, modified by recover.
func named() (r int) {
	defer func() {
		if recover() != nil {
			r = 42
		}
	}()
	x := val(5)
	defer func() { println("named sees", x) }()
	panic("boom")
}

// Method values: pointer receiver mutates the variable.
func methodValue() int {
	n := N(val(1))
	inc := n.Inc
	f := func() int { return int(n) }
	inc()
	inc()
	return f() // must be 3
}

// Value receiver method value copies at creation time.
func methodValueCopy() int {
	n := N(val(7))
	g := n.Get
	h := func() int { return g() + int(n) }
	return h()
}

// Variable whose address escapes through a pointer outside the closure.
func addrTaken() int {
	x := val(1)
	gp = &x
	f := func() int { return x }
	*gp = 9
	return f() // must be 9
}

// Mutation inside a second closure.
func twoClosures() int {
	x := val(1)
	r := func() int { return x }
	w := func() { x = 100 }
	w()
	return r() // must be 100
}

// Nested capture.
func nested() int {
	x := val(4)
	outer := func() func() int {
		return func() int { return x }
	}
	x2 := outer()
	return x2()
}

// Pointer-shaped values and GC: strings, slices, maps, interfaces, funcs, chans.
func pointers() {
	for i := 0; i < 20; i++ {
		s := string(rune('a' + i%26))
		sl := make([]int, 10)
		sl[9] = i
		m := map[int]int{i: i * 2}
		var e error = errors.New(s)
		fn := func() int { return i }
		ch := make(chan int, 1)
		ch <- i
		p := &P{i, i}
		keep = append(keep, func() int {
			return len(s) + sl[9] + m[i] + len(e.Error()) + fn() + <-ch + p.a
		})
	}
	runtime.GC()
	junk := make([][]byte, 0)
	for i := 0; i < 100; i++ {
		junk = append(junk, make([]byte, 64))
	}
	_ = junk
	runtime.GC()
	sum := 0
	for _, f := range keep {
		sum += f()
	}
	println("pointers", sum)
}

// Loop with goroutines reading by-value captures.
func goroutines() int {
	ch := make(chan int)
	for i := 0; i < 10; i++ {
		v := val(i)
		go func() { ch <- v * v }()
	}
	s := 0
	for i := 0; i < 10; i++ {
		s += <-ch
	}
	return s
}

// Generic with aggregate and non-aggregate instances.
func gen[T any](v T) func() T {
	x := v
	return func() T { return x }
}

// Same variable bound into a closure that is called both deferred and directly.
func deferAndCall() (out int) {
	x := val(6)
	f := func() int { return x * 2 }
	defer func() { out += f() }()
	return f()
}

// Closure defined in a block revisited via goto.
func gotoLoop() int {
	var fs []func() int
	i := 0
L:
	y := val(i)
	fs = append(fs, func() int { return y })
	i++
	if i < 3 {
		goto L
	}
	s := 0
	for _, f := range fs {
		s = s*10 + f()
	}
	return s // 12
}

// Range-over-func: the loop body is a synthesized closure.
func seq(yield func(int) bool) {
	for i := 0; i < 4; i++ {
		if !yield(i) {
			return
		}
	}
}

func rangeFunc() int {
	k := val(10)
	s := 0
	for v := range seq {
		s += v * k
	}
	return s
}

// Panic between the allocation and the store, recovered, in a loop.
func panicMid() (res int) {
	for i := 0; i < 3; i++ {
		func() {
			defer func() { recover() }()
			x := mayPanic(i)
			res += func() int { return x }()
		}()
	}
	return
}

func mayPanic(i int) int {
	if i == 1 {
		panic("mid")
	}
	return i + 1
}

// A single captured pointer is passed as the context itself. A closure that
// keeps it must keep the object alive and distinct for every iteration.
var kept []*P
var callbacks = []func(){func() {}}

//go:noinline
func keepContext(i int) {
	p := &P{i, i * 10}
	f := func() { kept = append(kept, p) }
	callbacks = append(callbacks, f)
	callbacks[len(callbacks)-1]()
}

//go:noinline
func clobber() {
	var buf [64]int
	for i := range buf {
		buf[i] = -1
	}
	sink = buf[len(buf)-1]
}

var sink int

func keptPointers() {
	for i := 0; i < 3; i++ {
		keepContext(i)
		clobber()
	}
	runtime.GC()
	for _, p := range kept {
		println("kept", p.a, p.b)
	}
}

func main() {
	println(recursive(), branch(true), branch(false), named())
	println(methodValue(), methodValueCopy(), addrTaken(), twoClosures(), nested())
	pointers()
	println(goroutines(), gen(1)(), gen("s")(), gen(P{1, 2})().b, gen([3]int{1, 2, 3})()[2])
	println(deferAndCall(), gotoLoop(), rangeFunc(), panicMid())
	keptPointers()
}

package main

// Tests the by-pointer passing convention for large aggregate parameters:
// value semantics (callee mutations and later caller mutations are
// invisible), all call paths (direct, func value, interface, defer,
// goroutine), and GC safety of the spilled copy's pointer fields.

import "runtime"

type big struct {
	p1, p2, p3, p4         *int
	a, b, c, d, e, f, g, h int
	i, j, k, l, m          int
}

func newInt(v int) *int {
	x := new(int)
	*x = v
	return x
}

func makeBig() big {
	return big{
		p1: newInt(1), p2: newInt(2), p3: newInt(3), p4: newInt(4),
		a: 5, b: 6, c: 7, d: 8, e: 9, f: 10, g: 11, h: 12,
		i: 13, j: 14, k: 15, l: 16, m: 17,
	}
}

// sink keeps allocations reachable so the optimizer cannot remove them.
var sink []byte

// sumBig forces allocation churn and collections during the call, so any
// missed root for the spilled copy's pointer fields would be exposed.
func sumBig(v big) int {
	allocSize := 4096
	if ^uintptr(0) <= 0xffff {
		// 16-bit and lower devices, such as AVR, where heap size is a real
		// constraint (same scaling as gc.go).
		allocSize = 64
	}
	for n := 0; n < 8; n++ {
		sink = make([]byte, allocSize)
		sink[n] = byte(n)
		runtime.GC()
	}
	return *v.p1 + *v.p2 + *v.p3 + *v.p4 + v.a + v.m
}

// mutate must not affect the caller's value.
//
//go:noinline
func mutate(v big) {
	v.a = 9999
	v.m = 9999
}

func (v big) sum() int {
	return sumBig(v)
}

type summer interface {
	sum() int
}

// small is a second summer implementation so the interface call below cannot
// be devirtualized: the invoke-wrapper spill path must be exercised.
type small struct{ v int }

func (s small) sum() int { return s.v }

var deferResult int

func storeSum(v big) {
	deferResult = sumBig(v)
}

// pickSum hides sumBig behind a function value so the call through f in main
// cannot be resolved to a static callee and must take the func-value path.
//
//go:noinline
func pickSum() func(big) int {
	return sumBig
}

func viaDefer(v big) {
	defer storeSum(v)
	v.m = 9999 // deferred call must see the value as it was at defer time
}

// viaGoroutine returns before the goroutine runs: a stack-spilled argument
// would dangle, so this pins the heap spill.
func viaGoroutine(v big, ch chan int) {
	go sendSum(v, ch)
}

func sendSum(v big, ch chan int) {
	ch <- sumBig(v)
}

func sumArray(a [20]int) int {
	total := 0
	for _, v := range a {
		total += v
	}
	return total
}

func main() {
	v := makeBig()
	println("direct:", sumBig(v)) // 1+2+3+4+5+17 = 32

	mutate(v)
	println("after mutate:", v.a, v.m) // 5 17

	f := pickSum()
	println("funcvalue:", f(v)) // 32

	summers := []summer{v, small{v: 100}}
	for _, s := range summers {
		println("iface:", s.sum()) // 32, then 100
	}

	viaDefer(v)
	println("defer:", deferResult) // 32

	ch := make(chan int)
	viaGoroutine(v, ch)
	println("goroutine:", <-ch) // 32

	var a [20]int
	for i := range a {
		a[i] = i
	}
	println("array:", sumArray(a)) // 190
}

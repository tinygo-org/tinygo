package main

// Loops that zero a slice are lowered to a memset. Check that the result is
// the same as running the loop.

type rasterizer struct {
	a []float32
}

func (z *rasterizer) reset() {
	for i := range z.a {
		z.a[i] = 0
	}
}

func resetRangeInt(z *rasterizer) {
	for i := range len(z.a) {
		z.a[i] = 0
	}
}

func resetFor(z *rasterizer) {
	for i := 0; i < len(z.a); i++ {
		z.a[i] = 0
	}
}

var global []int

//go:noinline
func clearGlobal() {
	for i := range global {
		global[i] = 0
	}
}

//go:noinline
func clearInts(s []int) {
	for i := range s {
		s[i] = 0
	}
}

//go:noinline
func clearFloats(s []float64) {
	for i := 0; i < len(s); i++ {
		s[i] = 0
	}
}

//go:noinline
func clearPointers(s []*int) {
	for i := range len(s) {
		s[i] = nil
	}
}

//go:noinline
func clearBytes(s []byte) {
	for i := range s {
		s[i] = 0
	}
}

type record struct {
	n    int
	name string
	v    any
	f    float64
}

//go:noinline
func clearRecords(s []record) {
	for i := range s {
		s[i] = record{}
	}
}

//go:noinline
func clearComplex(s []complex128) {
	for i := range s {
		s[i] = 0
	}
}

//go:noinline
func clearStrings(s []string) {
	for i := 0; i < len(s); i++ {
		s[i] = ""
	}
}

// The stores overwrite the pointer used to reach the slice. Optimized gc and
// TinyGo clear it all without the nil panic the loop would hit.
type selfRef struct{ a []*selfRef }

//go:noinline
func clearVia(w **selfRef) {
	for i := range (*w).a {
		(*w).a[i] = nil
	}
}

func main() {
	// Nil and empty slices.
	clearInts(nil)
	clearInts([]int{})
	println("nil and empty ok")

	// Neighbors of a sub-slice must be unchanged.
	backing := []int{1, 2, 3, 4, 5, 6}
	clearInts(backing[2:4])
	println("subslice:", backing[0], backing[1], backing[2], backing[3], backing[4], backing[5])

	// A zero-length sub-slice with capacity left over.
	clearInts(backing[1:1])
	println("empty subslice:", backing[1])

	bytes := []byte("hello world")
	clearBytes(bytes[6:])
	println("bytes:", string(bytes[:5]), bytes[6], bytes[10])

	floats := []float64{1.5, -2.5, 3.25}
	clearFloats(floats)
	println("floats:", floats[0] == 0, floats[1] == 0, floats[2] == 0, 1/floats[1] > 0)

	x, y := 1, 2
	ptrs := []*int{&x, &y, &x}
	clearPointers(ptrs)
	println("pointers:", ptrs[0] == nil, ptrs[1] == nil, ptrs[2] == nil)

	// A slice that aliases another variable.
	alias := backing[4:]
	clearInts(alias)
	println("alias:", backing[3], backing[4], backing[5])

	global = []int{7, 8, 9}
	clearGlobal()
	println("global:", global[0], global[1], global[2])

	// Methods and fields through a pointer.
	full := []float32{1, 2, 3, 4, 5}
	z := &rasterizer{a: full[1:3]}
	z.reset()
	println("reset:", int(full[0]), int(full[1]), int(full[2]), int(full[3]))
	full[1], full[2] = 9, 9
	resetRangeInt(z)
	println("rangeint:", int(full[0]), int(full[1]), int(full[2]), int(full[3]))
	full[1], full[2] = 9, 9
	resetFor(z)
	println("for:", int(full[0]), int(full[1]), int(full[2]), int(full[3]))

	// Empty receiver slice.
	(&rasterizer{}).reset()
	println("empty receiver ok")

	recs := []record{{1, "a", 2, 3}, {4, "b", "c", 5}, {6, "d", nil, 7}}
	clearRecords(recs[:2])
	println("records:", recs[0].n, recs[0].name == "", recs[0].v == nil, int(recs[1].f), recs[2].n, recs[2].name)

	cs := []complex128{1 + 2i, 3 + 4i}
	clearComplex(cs)
	println("complex:", cs[0] == 0, cs[1] == 0)

	strs := []string{"x", "y", "z"}
	clearStrings(strs[1:])
	println("strings:", strs[0], strs[1] == "", strs[2] == "", len(strs[2]))

	sr := make([]*selfRef, 2)
	sr[0] = &selfRef{a: sr}
	sr[1] = sr[0]
	func() {
		defer func() {
			println("self-referential: recovered:", recover() != nil, "cleared:", sr[0] == nil && sr[1] == nil)
		}()
		clearVia(&sr[0])
	}()

	// A nil receiver must still panic.
	for _, f := range []func(*rasterizer){(*rasterizer).reset, resetRangeInt, resetFor} {
		func() {
			defer func() {
				println("recovered:", recover() != nil)
			}()
			f(nil)
		}()
	}
}

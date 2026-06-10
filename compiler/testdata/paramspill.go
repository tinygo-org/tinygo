package main

// Aggregate parameters with more than 16 flattened scalar leaves are passed
// by pointer to a caller-owned copy in the Go ABI (see paramNeedsSpill), so
// backends that flatten aggregates (wasm) don't create functions with
// enormous parameter lists. Exported (C ABI) functions are not affected.

type big struct { // 17 leaves: passed by pointer
	a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p, q int
}

type edge struct { // 16 leaves: still passed by value
	a, b, c, d, e, f, g, h, i, j, k, l, m, n, o, p int
}

type withArray struct { // 1 + 16 = 17 leaves (arrays count): passed by pointer
	tag int
	buf [16]int32
}

func (b big) sum() int {
	return b.a + b.q
}

type summer interface {
	sum() int
}

var sink int

func takeBig(b big) int {
	return b.q
}

func takeEdge(e edge) int {
	return e.p
}

func takeArray(a [17]int32) int32 {
	return a[16]
}

func takeWithArray(w withArray) int32 {
	return w.buf[0]
}

//export takeBigC
func takeBigC(b big) int {
	return b.a
}

func spawnBig(b big) {
	sink = b.a
}

// pickTakeBig hides the callee behind a function value: //go:noinline keeps
// the SSA builder from resolving f(b) below to a static callee, so the call
// goes through the func-value decode path (extract the code pointer, nil
// check, indirect call) with the spilled parameter.
//
//go:noinline
func pickTakeBig() func(big) int {
	return takeBig
}

func callEverything(b big, e edge, a [17]int32, w withArray, s summer) int {
	sum := takeBig(b)
	sum += takeEdge(e)
	sum += int(takeArray(a))
	sum += int(takeWithArray(w))
	sum += takeBigC(b)
	f := pickTakeBig()
	sum += f(b)
	sum += s.sum()
	// Note: the lowering of this `go` statement depends on the default
	// scheduler of the wasm test target (asyncify).
	go spawnBig(b)
	return sum
}

func makeInterface(b big) summer {
	return b
}

// Exported method: the C ABI is used for the receiver, also in the interface
// invoke wrapper, so the >16-leaf receiver is not spilled.
//
//export sumWithArrayC
func (w withArray) sum() int {
	return w.tag
}

func makeInterfaceWithArray(w withArray) summer {
	return w
}

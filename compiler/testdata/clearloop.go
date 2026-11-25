package main

// Loops that zero every element of a slice are lowered to a memset, like
// clear(slice).

type rasterizer struct {
	a []float32
	b []float32
}

var global []int

// Form 1: range over the slice.

func (z *rasterizer) rangeField() {
	for i := range z.a {
		z.a[i] = 0
	}
}

func rangeGlobal() {
	for i := range global {
		global[i] = 0
	}
}

func rangeLocal(s []*int) {
	for i := range s {
		s[i] = nil
	}
}

func rangeBool(s []bool) {
	for i := range s {
		s[i] = false
	}
}

// Form 2: range over len(slice).

func rangeIntField(z *rasterizer) {
	for i := range len(z.a) {
		z.a[i] = 0
	}
}

// Form 3: three-clause loop.

func forField(z *rasterizer) {
	for i := 0; i < len(z.a); i++ {
		z.a[i] = 0
	}
}

func forLocal(s []uint16) {
	for i := 0; i < len(s); i++ {
		s[i] = 0
	}
}

// Zero values of other types.

type pair struct {
	a int
	s string
	e any
}

func structZero(s []pair) {
	for i := range s {
		s[i] = pair{}
	}
}

func structZeroVar(s []pair) {
	var z pair
	for i := range s {
		s[i] = z
	}
}

func arrayZero(s [][4]uint32) {
	for i := range s {
		s[i] = [4]uint32{}
	}
}

func complexZero(s []complex128) {
	for i := range s {
		s[i] = 0
	}
}

func stringZero(s []string) {
	for i := range s {
		s[i] = ""
	}
}

func interfaceZero(s []any) {
	for i := range s {
		s[i] = nil
	}
}

// The following loops must stay loops.

func nonZero(s []int) {
	for i := range s {
		s[i] = 1
	}
}

func extraStatement(s []int, t []int) {
	for i := range s {
		s[i] = 0
		t[0]++
	}
}

func getSlice() []int

func fromCall() {
	for i := range getSlice() {
		getSlice()[i] = 0
	}
}

func differentSlices(z *rasterizer) {
	for i := range z.a {
		z.b[i] = 0
	}
}

func indexUsedAfter(s []int) int {
	var i int
	for i = range s {
		s[i] = 0
	}
	return i
}

func stepTwo(s []int) {
	for i := 0; i < len(s); i += 2 {
		s[i] = 0
	}
}

type empty struct{}

func zeroSize(s []empty) {
	for i := range s {
		s[i] = empty{}
	}
}

func explicitFields(s []pair) {
	for i := range s {
		s[i] = pair{a: 0}
	}
}

func storeBeforeLoop(p *[]int, q *[]int) {
	s := *p
	*p = *q
	for i := range s {
		(*p)[i] = 0
	}
}

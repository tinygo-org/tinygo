package main

// Struct fields are compared in source order and array elements in index
// order, stopping at the first mismatch.
// https://go.dev/ref/spec#Comparison_operators

type smallIntFirst struct {
	A int
	X any
}

type smallItfFirst struct {
	X any
	A int
}

type largeIntFirst struct {
	A, B, C, D, E int
	X             any
}

type largeItfFirst struct {
	X             any
	A, B, C, D, E int
}

type stringFirst struct {
	S          string
	X          any
	A, B, C, D int
}

type twoItf struct {
	X, Y any
}

type nested struct {
	Inner struct{ A int }
	X     any
}

type elem struct {
	A int
	X any
}

var sl = []int{}

var (
	smallIntFirstX = smallIntFirst{A: 1, X: sl}
	smallIntFirstY = smallIntFirst{A: 2, X: sl}
	smallItfFirstX = smallItfFirst{X: sl, A: 1}
	smallItfFirstY = smallItfFirst{X: sl, A: 2}
	largeIntFirstX = largeIntFirst{A: 1, X: sl}
	largeIntFirstY = largeIntFirst{A: 2, X: sl}
	largeItfFirstX = largeItfFirst{X: sl, A: 1}
	largeItfFirstY = largeItfFirst{X: sl, A: 2}
	stringFirstX   = stringFirst{S: "a", X: sl}
	stringFirstY   = stringFirst{S: "b", X: sl}
	twoItfX        = twoItf{X: 1, Y: sl}
	twoItfY        = twoItf{X: 2, Y: sl}
	nestedX        = nested{X: sl}
	nestedY        = nested{X: sl}
	smallArrX      = [2]any{1, sl}
	smallArrY      = [2]any{2, sl}
	smallArrPanicX = [2]any{sl, 1}
	smallArrPanicY = [2]any{sl, 2}
	largeArrX      = [6]any{1, sl}
	largeArrY      = [6]any{2, sl}
	largeArrPanicX = [6]any{0, 0, 0, 0, sl, 1}
	largeArrPanicY = [6]any{0, 0, 0, 0, sl, 2}
	structArrX     = [3]elem{{A: 1}, {X: sl}}
	structArrY     = [3]elem{{A: 2}, {X: sl}}
)

func main() {
	nestedX.Inner.A = 1

	try("small struct, int then interface", func() bool { return smallIntFirstX == smallIntFirstY })
	try("small struct, interface then int", func() bool { return smallItfFirstX == smallItfFirstY })
	try("large struct, int then interface", func() bool { return largeIntFirstX == largeIntFirstY })
	try("large struct, interface then int", func() bool { return largeItfFirstX == largeItfFirstY })
	try("string then interface", func() bool { return stringFirstX == stringFirstY })
	try("interface then interface", func() bool { return twoItfX == twoItfY })
	try("nested struct then interface", func() bool { return nestedX == nestedY })
	try("small array", func() bool { return smallArrX == smallArrY })
	try("small array, panic first", func() bool { return smallArrPanicX == smallArrPanicY })
	try("large array", func() bool { return largeArrX == largeArrY })
	try("large array, panic first", func() bool { return largeArrPanicX == largeArrPanicY })
	try("array of structs", func() bool { return structArrX == structArrY })
	try("not equal", func() bool { return largeIntFirstX != largeIntFirstY })
}

func try(name string, f func() bool) {
	defer func() {
		if recover() != nil {
			println(name+":", "panic")
		}
	}()
	println(name+":", f())
}

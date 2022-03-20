// A Bits type and examples of operations to be performed on it.
//
// It's assignment semantics are just like a copy. If an independant copy
// is desired because internals will be manipulated unsafely, use Copy.
//
//	Empty()							Bits
//	Zero()							Bits
//	One()							Bits
//	Copy(a Bits)					Bits
//	Concat2(a, b Bits)				Bits
//	Double(a Bits)					Bits
//	Concat(a Bits, list ...Bits)	Bits
//	By8(a Bits)						Bits

package compiler

import (
	"fmt"
	"math/big"
)

// Output a to stdout with fmt.Println.
func o(a Bits) {
	fmt.Println(a.DebugString())
}

func ExampleBitsEmpty() {
	o(Empty())

	// Output:
	// (0, "")
}

func ExampleBitsZeroBitLayout() {
	o(Zero())

	// Output:
	// (1, "0")
}

func ExampleBitsOneBitLayout() {
	o(One())

	// Output:
	// (1, "1")
}

func ExampleBitsCopy() {
	x := Zero()
	y := Copy(x)
	o(x)
	o(y)
	// Unsafely manipulate x and then output x and y again
	// and show that y hasn't changed.
	x.bi.Set(big.NewInt(1))
	o(x)
	o(y)

	// Output:
	// (1, "0")
	// (1, "0")
	// (1, "1")
	// (1, "0")
}

func ExampleBitsDoubleOne() {
	o(Double(One()))

	// Output:
	// (2, "11")
}

func ExampleBitsDoubleZero() {
	o(Double(Zero()))

	// Output:
	// (2, "00")
}

func ExampleBitsConcat4() {
	zero := Zero()
	one := One()
	o(Concat(zero, one, zero, one)) // The first zero remains the lsb of the result.

	// Output:
	// (4, "1010")
}

func ExampleBitsConcat4Again() {
	zero := Zero()
	one := One()
	o(Concat(one, zero, zero, zero)) // The first one remains the lsb of the result.

	// Output:
	// (4, "0001")
}

func ExampleBitsConcatWithEmpties() {
	zero := Zero()
	one := One()
	empty := Empty()
	o(Concat(one, zero, zero, zero))        // 0001
	o(Concat(empty, one, zero, zero, zero)) // 0001 (no difference)
	o(Concat(one, zero, empty, zero, zero)) // 0001 (no difference)

	// Output:
	// (4, "0001")
	// (4, "0001")
	// (4, "0001")
}

func ExampleBitsBy8_Zero() {
	zero := Zero()
	o(By8(zero))

	// Output:
	// (8, "00000000")
}

func ExampleBitsBy8_One() {
	one := One()
	o(By8(one))

	// Output:
	// (8, "11111111")
}

func ExampleBitsBy8_Empty() {
	empty := Empty()
	o(By8(empty))

	// Output:
	// (0, "")
}

func ExampleBitsFullBlown() {
	a := Concat(Zero(), Zero(), One())        // "100"
	b := Concat(One(), One(), Zero(), Zero()) // "0011"
	o(a)
	o(b)
	o(Concat(By8(a), By8(b)))

	// Output:
	// (3, "100")
	// (4, "0011")
	// (56, "00110011001100110011001100110011100100100100100100100100")
}

func ExampleBitsFullBlownButFirstEmpty() {
	a := Empty()                              // ""
	b := Concat(One(), One(), Zero(), Zero()) // "0011"
	o(a)
	o(b)
	o(Concat(By8(a), By8(b)))

	// Output:
	// (0, "")
	// (4, "0011")
	// (32, "00110011001100110011001100110011")
}

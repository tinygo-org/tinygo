package compiler

import (
	"fmt"
	"math/big"
	"strings"
)

// Bits tracks the number of bits in a bitmap along with the bits themselves
// represented by a big.Int.
type Bits struct {
	size uint     // How many bits are being represented. Can also be considered its length.
	bi   *big.Int // The bits being represented.
}

func (a Bits) AssertValidBitLen() {
	if uint(a.bi.BitLen()) > a.size {
		panic(fmt.Sprintf("a.bi.BitLen() > a.size, %d, %d", a.bi.BitLen(), a.size))
	}
}

// Returns the size, the number of bits being represented.
func (a Bits) Size() uint {
	return a.size
}

// Returns the bits themselves. Normally the bits without the size isn't
// very useful.
func (a Bits) Int() *big.Int {
	return new(big.Int).Set(a.bi)
}

// Returns the string of zeros and ones of the binary representation of the bits.
func (a Bits) String() string {
	if a.size == 0 {
		return ""
	}
	s := a.bi.Text(2)
	// The binary representation should not be longer than
	// the size being tracked.
	if len(s) > int(a.size) {
		panic("malformed Bits")
	}
	if len(s) == int(a.size) {
		return s
	}
	// Pad front with missing zeros.
	return strings.Repeat("0", int(a.size)-len(s)) + s
}

// Return a string, almost a GoString style string (only a little prettier),
// that should aid in debugging a Bits's contents. The examples use this.
func (a Bits) DebugString() string {
	return fmt.Sprintf("(%d, %q)", a.size, a.String())
}

// Return an empty Bits. It has no width. It represents 0 bits
// although it does have a big.Int instance of zero.
func Empty() Bits {
	return Bits{
		size: 0,
		bi:   big.NewInt(0),
	}
}

// Return the Bits for one bit set to 0.
func Zero() Bits {
	return Bits{
		size: 1,
		bi:   big.NewInt(0),
	}
}

// Return the Bits for one bit set to 1.
func One() Bits {
	return Bits{
		size: 1,
		bi:   big.NewInt(1),
	}
}

// Copy the Bits. Normally not necessary but to preserve a copy before
// manipulating the *big.Int within one, it comes in handy.
func Copy(a Bits) Bits {
	return Bits{
		a.size,
		new(big.Int).Set(a.bi),
	}
}

// Return the concatentation of a and b. The semantics are that the bits of a
// remain in the same position of the returned value. The bits of b
// are left-shifted by the size of a.
func Concat2(a, b Bits) Bits {
	n := new(big.Int).Lsh(b.bi, a.size)
	return Bits{
		a.size + b.size,
		n.Or(a.bi, n),
	}
}

func Double(a Bits) Bits {
	return Concat2(a, a)
}

// Return a repeated n times.
func Repeat(a Bits, n uint) Bits {
	if n == 0 {
		return Empty()
	}
	r := a
	for i := uint(1); i < n; i++ {
		r = Concat2(r, a)
	}
	return r
}

// Concat is like the append builtin; those appearing
// in list are appended to a. Unlike append, there is no backing store
// that might be shared. The lsb of the returned value will be the same
// bits found in a, i.e. the bits in a are not shifted, it is the bits
// in the list elements that are shifted to be bitwise-ORed into the result.
func Concat(a Bits, list ...Bits) Bits {
	for _, i := range list {
		a = Concat2(a, i)
	}
	return a
}

// Return a repeated 8 times.
func By8(a Bits) Bits {
	// Double three times.
	// This uses three allocations when technically it could be done with one.
	return Double(Double(Double(a)))
}

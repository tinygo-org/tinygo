package runtime

// This file implements Go interfaces.
//
// Interfaces are represented as a pair of {typecode, value}, where value can be
// anything (including non-pointers).
//
// Signatures itself are not matched on strings, but on uniqued numbers that
// contain the name and the signature of the function (to save space), think of
// signatures as interned strings at compile time.
//
// The typecode is a small number unique for the Go type. All typecodes <
// firstInterfaceNum do not have any methods and typecodes >= firstInterfaceNum
// all have at least one method. This means that methodSetRanges does not need
// to contain types without methods and is thus indexed starting at a typecode
// with number firstInterfaceNum.
//
// To further conserve some space, the methodSetRange (as the name indicates)
// doesn't contain a list of methods and function pointers directly, but instead
// just indexes into methodSetSignatures and methodSetFunctions which contains
// the mapping from uniqued signature to function pointer.

type _interface struct {
	typecode uint16
	value    *uint8
}

// This struct indicates the range of methods in the methodSetSignatures and
// methodSetFunctions arrays that belong to this named type.
type methodSetRange struct {
	index  uint16 // start index into interfaceSignatures and interfaceFunctions
	length uint16 // number of methods
}

// Global constants that will be set by the compiler. The arrays are of size 0,
// which is a dummy value, but will be bigger after the compiler has filled them
// in.
var (
	firstInterfaceNum   uint16            // the lowest typecode that has at least one method
	methodSetRanges     [0]methodSetRange // indexes into methodSetSignatures and methodSetFunctions
	methodSetSignatures [0]uint16         // uniqued method ID
	methodSetFunctions  [0]*uint8         // function pointer of method
)

// Get the function pointer for the method on the interface.
// This is a compiler intrinsic.
//go:nobounds
func interfaceMethod(itf _interface, method uint16) *uint8 {
	// This function doesn't do bounds checking as the supplied method must be
	// in the list of signatures. The compiler will only emit
	// runtime.interfaceMethod calls when the method actually exists on this
	// interface (proven by the typechecker).
	i := methodSetRanges[itf.typecode-firstInterfaceNum].index
	for {
		if methodSetSignatures[i] == method {
			return methodSetFunctions[i]
		}
		i++
	}
}

// Return true iff both interfaces are equal.
func interfaceEqual(x, y _interface) bool {
	if x.typecode != y.typecode {
		// Different dynamic type so always unequal.
		return false
	}
	if x.typecode == 0 {
		// Both interfaces are nil, so they are equal.
		return true
	}
	// TODO: depends on reflection.
	panic("unimplemented: interface equality")
}

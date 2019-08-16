package runtime

// This file implements Go interfaces.
//
// Interfaces are represented as a pair of {typecode, value}, where value can be
// anything (including non-pointers).

import "unsafe"

type _interface struct {
	typecode uintptr
	value    unsafe.Pointer
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

// interfaceTypeAssert is called when a type assert without comma-ok still
// returns false.
func interfaceTypeAssert(ok bool) {
	if !ok {
		runtimePanic("type assert failed")
	}
}

// The following declarations are only used during IR construction. They are
// lowered to inline IR in the interface lowering pass.
// See compiler/interface-lowering.go for details.

type interfaceMethodInfo struct {
	signature *uint8  // external *i8 with a name identifying the Go function signature
	funcptr   uintptr // bitcast from the actual function pointer
}

type typecodeID struct {
	// Depending on the type kind of this typecodeID, this pointer is something
	// different:
	// * basic types: null
	// * named type: the underlying type
	// * interface: null
	// * chan/pointer/slice/array: the element type
	// * struct: bitcast of global with structField array
	// * func/map: TODO
	references *typecodeID

	// The array length, for array types.
	length uintptr
}

// structField is used by the compiler to pass information to the interface
// lowering pass. It is not used in the final binary.
type structField struct {
	typecode *typecodeID // type of this struct field
	name     *uint8      // pointer to char array
	tag      *uint8      // pointer to char array, or nil
	embedded bool
}

// Pseudo type used before interface lowering. By using a struct instead of a
// function call, this is simpler to reason about during init interpretation
// than a function call. Also, by keeping the method set around it is easier to
// implement interfaceImplements in the interp package.
type typeInInterface struct {
	typecode  *typecodeID
	methodSet *interfaceMethodInfo // nil or a GEP of an array
}

// Pseudo function call used during a type assert. It is used during interface
// lowering, to assign the lowest type numbers to the types with the most type
// asserts. Also, it is replaced with const false if this type assert can never
// happen.
func typeAssert(actualType uintptr, assertedType *typecodeID) bool

// Pseudo function call that returns whether a given type implements all methods
// of the given interface.
func interfaceImplements(typecode uintptr, interfaceMethodSet **uint8) bool

// Pseudo function that returns a function pointer to the method to call.
// See the interface lowering pass for how this is lowered to a real call.
func interfaceMethod(typecode uintptr, interfaceMethodSet **uint8, signature *uint8) uintptr

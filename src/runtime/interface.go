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
	signature *uint8 // external *i8 with a name identifying the Go function signature
	funcptr   *uint8 // bitcast from the actual function pointer
}

// Pseudo function call used while putting a concrete value in an interface,
// that must be lowered to a constant uintptr.
func makeInterface(typecode *uint8, methodSet *interfaceMethodInfo) uintptr

// Pseudo function call used during a type assert. It is used during interface
// lowering, to assign the lowest type numbers to the types with the most type
// asserts. Also, it is replaced with const false if this type assert can never
// happen.
func typeAssert(actualType uintptr, assertedType *uint8) bool

// Pseudo function call that returns whether a given type implements all methods
// of the given interface.
func interfaceImplements(typecode uintptr, interfaceMethodSet **uint8) bool

// Pseudo function that returns a function pointer to the method to call.
// See the interface lowering pass for how this is lowered to a real call.
func interfaceMethod(typecode uintptr, interfaceMethodSet **uint8, signature *uint8) *uint8

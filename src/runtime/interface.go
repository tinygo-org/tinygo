package runtime

// This file implements Go interfaces.
//
// Interfaces are represented as a pair of {typecode, value}, where value can be
// anything (including non-pointers).

import (
	"reflect"
	"unsafe"
)

type _interface struct {
	typecode uintptr
	value    unsafe.Pointer
}

//go:inline
func composeInterface(typecode uintptr, value unsafe.Pointer) _interface {
	return _interface{typecode, value}
}

//go:inline
func decomposeInterface(i _interface) (uintptr, unsafe.Pointer) {
	return i.typecode, i.value
}

// Return true iff both interfaces are equal.
func interfaceEqual(x, y interface{}) bool {
	return reflectValueEqual(reflect.ValueOf(x), reflect.ValueOf(y))
}

func reflectValueEqual(x, y reflect.Value) bool {
	if x.Type() == 0 || y.Type() == 0 {
		// One of them is nil.
		return x.Type() == y.Type()
	}

	if x.Type() != y.Type() {
		// The type is not the same, which means the interfaces are definitely
		// not the same.
		return false
	}

	switch x.Type().Kind() {
	case reflect.Bool:
		return x.Bool() == y.Bool()
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return x.Int() == y.Int()
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64, reflect.Uintptr:
		return x.Uint() == y.Uint()
	case reflect.Float32, reflect.Float64:
		return x.Float() == y.Float()
	case reflect.Complex64, reflect.Complex128:
		return x.Complex() == y.Complex()
	case reflect.String:
		return x.String() == y.String()
	case reflect.Chan, reflect.Ptr, reflect.UnsafePointer:
		return x.Pointer() == y.Pointer()
	case reflect.Array:
		for i := 0; i < x.Len(); i++ {
			if !reflectValueEqual(x.Index(i), y.Index(i)) {
				return false
			}
		}
		return true
	case reflect.Struct:
		for i := 0; i < x.NumField(); i++ {
			if !reflectValueEqual(x.Field(i), y.Field(i)) {
				return false
			}
		}
		return true
	default:
		runtimePanic("comparing un-comparable type")
		return false // unreachable
	}
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
	typecode  *typecodeID          // element type, underlying type, or reference to struct fields
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

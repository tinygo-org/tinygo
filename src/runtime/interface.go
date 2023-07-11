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
	typecode unsafe.Pointer
	value    unsafe.Pointer
}

//go:inline
func composeInterface(typecode, value unsafe.Pointer) _interface {
	return _interface{typecode, value}
}

//go:inline
func decomposeInterface(i _interface) (unsafe.Pointer, unsafe.Pointer) {
	return i.typecode, i.value
}

// Return true iff both interfaces are equal.
func interfaceEqual(x, y interface{}) bool {
	return reflectValueEqual(reflect.ValueOf(x), reflect.ValueOf(y))
}

func reflectValueEqual(x, y reflect.Value) bool {
	// Note: doing a x.Type() == y.Type() comparison would not work here as that
	// would introduce an infinite recursion: comparing two reflect.Type values
	// is done with this reflectValueEqual runtime call.
	if x.RawType() == nil || y.RawType() == nil {
		// One of them is nil.
		return x.RawType() == y.RawType()
	}

	if x.RawType() != y.RawType() {
		// The type is not the same, which means the interfaces are definitely
		// not the same.
		return false
	}

	switch x.RawType().Kind() {
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
		return x.UnsafePointer() == y.UnsafePointer()
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
	case reflect.Interface:
		return reflectValueEqual(x.Elem(), y.Elem())
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

type structField struct {
	typecode unsafe.Pointer // type of this struct field
	data     *uint8         // pointer to byte array containing name, tag, varint-encoded offset, and some flags
}

// Pseudo function call used during a type assert. It is used during interface
// lowering, to assign the lowest type numbers to the types with the most type
// asserts. Also, it is replaced with const false if this type assert can never
// happen.
func typeAssert(actualType unsafe.Pointer, assertedType *uint8) bool

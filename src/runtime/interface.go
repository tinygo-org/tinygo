package runtime

// This file implements Go interfaces.
//
// Interfaces are represented as a pair of {typecode, value}, where value can be
// anything (including non-pointers).

import (
	"internal/reflectlite"
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
	return reflectValueEqual(reflectlite.ValueOf(x), reflectlite.ValueOf(y))
}

func reflectValueEqual(x, y reflectlite.Value) bool {
	// Note: doing a x.Type() == y.Type() comparison would not work here as that
	// would introduce an infinite recursion: comparing two reflectlite.Type values
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
	case reflectlite.Bool:
		return x.Bool() == y.Bool()
	case reflectlite.Int, reflectlite.Int8, reflectlite.Int16, reflectlite.Int32, reflectlite.Int64:
		return x.Int() == y.Int()
	case reflectlite.Uint, reflectlite.Uint8, reflectlite.Uint16, reflectlite.Uint32, reflectlite.Uint64, reflectlite.Uintptr:
		return x.Uint() == y.Uint()
	case reflectlite.Float32, reflectlite.Float64:
		return x.Float() == y.Float()
	case reflectlite.Complex64, reflectlite.Complex128:
		return x.Complex() == y.Complex()
	case reflectlite.String:
		return x.String() == y.String()
	case reflectlite.Chan, reflectlite.Ptr, reflectlite.UnsafePointer:
		return x.UnsafePointer() == y.UnsafePointer()
	case reflectlite.Array:
		for i := 0; i < x.Len(); i++ {
			if !reflectValueEqual(x.Index(i), y.Index(i)) {
				return false
			}
		}
		return true
	case reflectlite.Struct:
		for i := 0; i < x.NumField(); i++ {
			if !reflectValueEqual(x.Field(i), y.Field(i)) {
				return false
			}
		}
		return true
	case reflectlite.Interface:
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

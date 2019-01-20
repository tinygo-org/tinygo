package reflect

import (
	"unsafe"
)

// The compiler uses a compact encoding to store type information. Unlike the
// main Go compiler, most of the types are stored directly in the type code.
//
// Type code bit allocation:
// xxxxx0: basic types, where xxxxx is the basic type number (never 0).
//         The higher bits indicate the named type, if any.
//  nxxx1: complex types, where n indicates whether this is a named type (named
//         if set) and xxx contains the type kind number:
//             0 (0001): Chan
//             1 (0011): Interface
//             2 (0101): Ptr
//             3 (0111): Slice
//             4 (1001): Array
//             5 (1011): Func
//             6 (1101): Map
//             7 (1111): Struct
//         The higher bits are either the contents of the type depending on the
//         type (if n is clear) or indicate the number of the named type (if n
//         is set).

type Kind uintptr

// Copied from reflect/type.go
// https://golang.org/src/reflect/type.go?s=8302:8316#L217
const (
	Invalid Kind = iota
	Bool
	Int
	Int8
	Int16
	Int32
	Int64
	Uint
	Uint8
	Uint16
	Uint32
	Uint64
	Uintptr
	Float32
	Float64
	Complex64
	Complex128
	String
	UnsafePointer
	Chan
	Interface
	Ptr
	Slice
	Array
	Func
	Map
	Struct
)

func (k Kind) String() string {
	switch k {
	case Bool:
		return "bool"
	case Int:
		return "int"
	case Int8:
		return "int8"
	case Int16:
		return "int16"
	case Int32:
		return "int32"
	case Int64:
		return "int64"
	case Uint:
		return "uint"
	case Uint8:
		return "uint8"
	case Uint16:
		return "uint16"
	case Uint32:
		return "uint32"
	case Uint64:
		return "uint64"
	case Uintptr:
		return "uintptr"
	case Float32:
		return "float32"
	case Float64:
		return "float64"
	case Complex64:
		return "complex64"
	case Complex128:
		return "complex128"
	case String:
		return "string"
	case UnsafePointer:
		return "unsafe.Pointer"
	case Chan:
		return "chan"
	case Interface:
		return "interface"
	case Ptr:
		return "ptr"
	case Slice:
		return "slice"
	case Array:
		return "array"
	case Func:
		return "func"
	case Map:
		return "map"
	case Struct:
		return "struct"
	default:
		return "invalid"
	}
}

// basicType returns a new Type for this kind if Kind is a basic type.
func (k Kind) basicType() Type {
	return Type(k << 1)
}

// The typecode as used in an interface{}.
type Type uintptr

func TypeOf(i interface{}) Type {
	return ValueOf(i).typecode
}

func (t Type) String() string {
	return "T"
}

func (t Type) Kind() Kind {
	if t % 2 == 0 {
		// basic type
		return Kind((t >> 1) % 32)
	} else {
		return Kind(t >> 1) % 8 + 19
	}
}

func (t Type) Elem() Type {
	switch t.Kind() {
	case Chan, Ptr, Slice:
		if (t >> 4) % 2 != 0 {
			panic("unimplemented: (reflect.Type).Elem() for named types")
		}
		return t >> 5
	default: // not implemented: Array, Map
		panic("unimplemented: (reflect.Type).Elem()")
	}
}

func (t Type) Field(i int) StructField {
	panic("unimplemented: (reflect.Type).Field()")
}

func (t Type) Bits() int {
	panic("unimplemented: (reflect.Type).Bits()")
}

func (t Type) Len() int {
	panic("unimplemented: (reflect.Type).Len()")
}

func (t Type) NumField() int {
	panic("unimplemented: (reflect.Type).NumField()")
}

func (t Type) Size() uintptr {
	switch t.Kind() {
	case Bool, Int8, Uint8:
		return 1
	case Int16, Uint16:
		return 2
	case Int32, Uint32:
		return 4
	case Int64, Uint64:
		return 8
	case Int, Uint:
		return unsafe.Sizeof(int(0))
	case Uintptr:
		return unsafe.Sizeof(uintptr(0))
	case Float32:
		return 4
	case Float64:
		return 8
	case Complex64:
		return 8
	case Complex128:
		return 16
	case String:
		return unsafe.Sizeof(StringHeader{})
	case UnsafePointer, Chan, Map, Ptr:
		return unsafe.Sizeof(uintptr(0))
	case Slice:
		return unsafe.Sizeof(SliceHeader{})
	default:
		// Size unknown.
		return 0
	}
}

type StructField struct {
	Name string
	Type Type
}

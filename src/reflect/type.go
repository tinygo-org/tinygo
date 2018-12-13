package reflect

import (
	"unsafe"
)

// A Kind is the number that the compiler uses for this type.
//
// Not used directly. These types are all replaced with the number the compiler
// uses internally for the type.
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
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
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
	case Slice:
		return "slice"
	default:
		return "invalid"
	}
}

// basicType returns a new Type for this kind if Kind is a basic type.
func (k Kind) basicType() Type {
	return Type(k << 2 | 0)
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
	if t % 4 == 0 {
		// Basic type
		return Kind(t >> 2)
	} else if t % 4 == 1 {
		// Slice
		return Slice
	} else {
		return Invalid // TODO
	}
}

func (t Type) Elem() Type {
	switch t.Kind() {
	case Slice:
		return t >> 2
	default: // not implemented: Array, Chan, Map, Ptr
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
	case UnsafePointer:
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

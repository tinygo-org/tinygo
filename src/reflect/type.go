package reflect

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
	default:
		return "T"
	}
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
	if t & 1 == 0 {
		// Basic type
		return Kind(t >> 1)
	} else {
		return Invalid // TODO
	}
}

func (t Type) Elem() Type {
	panic("unimplemented: (reflect.Type).Elem()")
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
	panic("unimplemented: (reflect.Type).Size()")
}

type StructField struct {
	Name string
	Type Type
}

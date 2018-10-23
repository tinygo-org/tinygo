package reflect

// A Kind is the number that the compiler uses for this type.
//
// Not used directly. These types are all replaced with the number the compiler
// uses internally for the type.
type Kind uint16

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
	Array
	Chan
	Func
	Interface
	Map
	Ptr
	Slice
	String
	Struct
	UnsafePointer
)

// The typecode as used in an interface{}.
type Type uint16

func TypeOf(i interface{}) Type {
	return ValueOf(i).typecode
}

func (t Type) String() string {
	return "T"
}

func (t Type) Kind() Kind {
	return Invalid // TODO
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

type StructField struct {
	Name string
}

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
	if t%2 == 0 {
		// basic type
		return Kind((t >> 1) % 32)
	} else {
		return Kind(t>>1)%8 + 19
	}
}

// Elem returns the element type for channel, slice and array types, the
// pointed-to value for pointer types, and the key type for map types.
func (t Type) Elem() Type {
	switch t.Kind() {
	case Chan, Ptr, Slice:
		return t.stripPrefix()
	case Array:
		index := t.stripPrefix()
		elem, _ := readVarint(unsafe.Pointer(uintptr(unsafe.Pointer(&arrayTypesSidetable)) + uintptr(index)))
		return Type(elem)
	default: // not implemented: Map
		panic("unimplemented: (reflect.Type).Elem()")
	}
}

// stripPrefix removes the "prefix" (the first 5 bytes of the type code) from
// the type code. If this is a named type, it will resolve the underlying type
// (which is the data for this named type). If it is not, the lower bits are
// simply shifted off.
//
// The behavior is only defined for non-basic types.
func (t Type) stripPrefix() Type {
	// Look at the 'n' bit in the type code (see the top of this file) to see
	// whether this is a named type.
	if (t>>4)%2 != 0 {
		// This is a named type. The data is stored in a sidetable.
		namedTypeNum := t >> 5
		n := *(*uintptr)(unsafe.Pointer(uintptr(unsafe.Pointer(&namedNonBasicTypesSidetable)) + uintptr(namedTypeNum)*unsafe.Sizeof(uintptr(0))))
		return Type(n)
	}
	// Not a named type, so the value is stored directly in the type code.
	return t >> 5
}

// Field returns the type of the i'th field of this struct type. It panics if t
// is not a struct type.
func (t Type) Field(i int) StructField {
	if t.Kind() != Struct {
		panic(&TypeError{"Field"})
	}
	structIdentifier := t.stripPrefix()

	numField, p := readVarint(unsafe.Pointer(uintptr(unsafe.Pointer(&structTypesSidetable)) + uintptr(structIdentifier)))
	if uint(i) >= uint(numField) {
		panic("reflect: field index out of range")
	}

	// Iterate over every field in the struct and update the StructField each
	// time, until the target field has been reached. This is very much not
	// efficient, but it is easy to implement.
	// Adding a jump table at the start to jump to the field directly would
	// make this much faster, but that would also impact code size.
	field := StructField{}
	offset := uintptr(0)
	for fieldNum := 0; fieldNum <= i; fieldNum++ {
		// Read some flags of this field, like whether the field is an
		// embedded field.
		flagsByte := *(*uint8)(p)
		p = unsafe.Pointer(uintptr(p) + 1)

		// Read the type of this struct field.
		var fieldType uintptr
		fieldType, p = readVarint(p)
		field.Type = Type(fieldType)

		// Move Offset forward to align it to this field's alignment.
		// Assume alignment is a power of two.
		offset = align(offset, uintptr(field.Type.Align()))
		field.Offset = offset
		offset += field.Type.Size() // starting (unaligned) offset for next field

		// Read the field name.
		var nameNum uintptr
		nameNum, p = readVarint(p)
		field.Name = readStringSidetable(unsafe.Pointer(&structNamesSidetable), nameNum)

		// The first bit in the flagsByte indicates whether this is an embedded
		// field.
		field.Anonymous = flagsByte&1 != 0

		// The second bit indicates whether there is a tag.
		if flagsByte&2 != 0 {
			// There is a tag.
			var tagNum uintptr
			tagNum, p = readVarint(p)
			field.Tag = readStringSidetable(unsafe.Pointer(&structNamesSidetable), tagNum)
		} else {
			// There is no tag.
			field.Tag = ""
		}

		// The third bit indicates whether this field is exported.
		if flagsByte&4 != 0 {
			// This field is exported.
			field.PkgPath = ""
		} else {
			// This field is unexported.
			// TODO: list the real package path here. Storing it should not
			// significantly impact binary size as there is only a limited
			// number of packages in any program.
			field.PkgPath = "<unimplemented>"
		}
	}

	return field
}

// Bits returns the number of bits that this type uses. It is only valid for
// arithmetic types (integers, floats, and complex numbers). For other types, it
// will panic.
func (t Type) Bits() int {
	kind := t.Kind()
	if kind >= Int && kind <= Complex128 {
		return int(t.Size()) * 8
	}
	panic(TypeError{"Bits"})
}

// Len returns the number of elements in this array. It panics of the type kind
// is not Array.
func (t Type) Len() int {
	if t.Kind() != Array {
		panic(TypeError{"Len"})
	}

	// skip past the element type
	arrayIdentifier := t.stripPrefix()
	_, p := readVarint(unsafe.Pointer(uintptr(unsafe.Pointer(&arrayTypesSidetable)) + uintptr(arrayIdentifier)))

	// Read the array length.
	arrayLen, _ := readVarint(p)
	return int(arrayLen)
}

// NumField returns the number of fields of a struct type. It panics for other
// type kinds.
func (t Type) NumField() int {
	if t.Kind() != Struct {
		panic(&TypeError{"NumField"})
	}
	structIdentifier := t.stripPrefix()
	n, _ := readVarint(unsafe.Pointer(uintptr(unsafe.Pointer(&structTypesSidetable)) + uintptr(structIdentifier)))
	return int(n)
}

// Size returns the size in bytes of a given type. It is similar to
// unsafe.Sizeof.
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
	case Interface:
		return unsafe.Sizeof(interfaceHeader{})
	case Array:
		return t.Elem().Size() * uintptr(t.Len())
	case Struct:
		numField := t.NumField()
		if numField == 0 {
			return 0
		}
		lastField := t.Field(numField - 1)
		return lastField.Offset + lastField.Type.Size()
	default:
		panic("unimplemented: size of type")
	}
}

// Align returns the alignment of this type. It is similar to calling
// unsafe.Alignof.
func (t Type) Align() int {
	switch t.Kind() {
	case Bool, Int8, Uint8:
		return int(unsafe.Alignof(int8(0)))
	case Int16, Uint16:
		return int(unsafe.Alignof(int16(0)))
	case Int32, Uint32:
		return int(unsafe.Alignof(int32(0)))
	case Int64, Uint64:
		return int(unsafe.Alignof(int64(0)))
	case Int, Uint:
		return int(unsafe.Alignof(int(0)))
	case Uintptr:
		return int(unsafe.Alignof(uintptr(0)))
	case Float32:
		return int(unsafe.Alignof(float32(0)))
	case Float64:
		return int(unsafe.Alignof(float64(0)))
	case Complex64:
		return int(unsafe.Alignof(complex64(0)))
	case Complex128:
		return int(unsafe.Alignof(complex128(0)))
	case String:
		return int(unsafe.Alignof(StringHeader{}))
	case UnsafePointer, Chan, Map, Ptr:
		return int(unsafe.Alignof(uintptr(0)))
	case Slice:
		return int(unsafe.Alignof(SliceHeader{}))
	case Interface:
		return int(unsafe.Alignof(interfaceHeader{}))
	case Struct:
		numField := t.NumField()
		alignment := 1
		for i := 0; i < numField; i++ {
			fieldAlignment := t.Field(i).Type.Align()
			if fieldAlignment > alignment {
				alignment = fieldAlignment
			}
		}
		return alignment
	default:
		panic("unimplemented: alignment of type")
	}
}

// FieldAlign returns the alignment if this type is used in a struct field. It
// is currently an alias for Align() but this might change in the future.
func (t Type) FieldAlign() int {
	return t.Align()
}

// AssignableTo returns whether a value of type u can be assigned to a variable
// of type t.
func (t Type) AssignableTo(u Type) bool {
	if t == u {
		return true
	}
	if t.Kind() == Interface {
		panic("reflect: unimplemented: assigning to interface of different type")
	}
	return false
}

// Comparable returns whether values of this type can be compared to each other.
func (t Type) Comparable() bool {
	switch t.Kind() {
	case Bool, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
		return true
	case Float32, Float64, Complex64, Complex128:
		return true
	case String:
		return true
	case UnsafePointer:
		return true
	case Chan:
		return true
	case Interface:
		return true
	case Ptr:
		return true
	case Slice:
		return false
	case Array:
		return t.Elem().Comparable()
	case Func:
		return false
	case Map:
		return false
	case Struct:
		numField := t.NumField()
		for i := 0; i < numField; i++ {
			if !t.Field(i).Type.Comparable() {
				return false
			}
		}
		return true
	default:
		panic(TypeError{"Comparable"})
	}
}

// A StructField describes a single field in a struct.
type StructField struct {
	// Name indicates the field name.
	Name string

	// PkgPath is the package path where the struct containing this field is
	// declared for unexported fields, or the empty string for exported fields.
	PkgPath string

	Type      Type
	Tag       string
	Anonymous bool
	Offset    uintptr
}

// TypeError is the error that is used in a panic when invoking a method on a
// type that is not applicable to that type.
type TypeError struct {
	Method string
}

func (e *TypeError) Error() string {
	return "reflect: call of reflect.Type." + e.Method + " on invalid type"
}

func align(offset uintptr, alignment uintptr) uintptr {
	return (offset + alignment - 1) &^ (alignment - 1)
}

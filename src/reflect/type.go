// Copyright 2009 The Go Authors. All rights reserved.
// Use of this source code is governed by a BSD-style
// license that can be found in the LICENSE file.

// Type information of an interface is stored as a pointer to a global in the
// interface type (runtime._interface). This is called a type struct.
// It always starts with a byte that contains both the type kind and a few
// flags. In most cases it also contains a pointer to another type struct
// (ptrTo), that is the pointer type of the current type (for example, type int
// also has a pointer to the type *int). The exception is pointer types, to
// avoid infinite recursion.
//
// The layouts specifically look like this:
// - basic types (Bool..UnsafePointer):
//     meta         uint8 // actually: kind + flags
//     ptrTo        *typeStruct
// - named types (see elemType):
//     meta         uint8
//     ptrTo        *typeStruct
//     underlying   *typeStruct // the underlying, non-named type
// - channels and slices (see elemType):
//     meta          uint8
//     ptrTo        *typeStruct
//     elementType  *typeStruct // the type that you get with .Elem()
// - pointer types (see ptrType, this doesn't include chan, map, etc):
//     meta         uint8
//     elementType  *typeStruct
// - array types (see arrayType)
//     meta         uint8
//     ptrTo        *typeStruct
//     elem         *typeStruct // element type of the array
//     arrayLen     uintptr     // length of the array (this is part of the type)
// - map types (this is still missing the key and element types)
//     meta         uint8
//     ptrTo        *typeStruct
//     elem         *typeStruct
//     key          *typeStruct
// - struct types (see structType):
//     meta         uint8
//     numField     uint16
//     ptrTo        *typeStruct
//     pkgpath      *byte
//     fields       [...]structField // the remaining fields are all of type structField
// - interface types (this is missing the interface methods):
//     meta         uint8
//     ptrTo        *typeStruct
// - signature types (this is missing input and output parameters):
//     meta         uint8
//     ptrTo        *typeStruct
// - named types
//     meta         uint8
//     ptrTo        *typeStruct
//     elem         *typeStruct // underlying type
//     pkg          *byte       // pkgpath
//     name         [1]byte     // actual name; length nlem
//
// The type struct is essentially a union of all the above types. Which it is,
// can be determined by looking at the meta byte.

package reflect

import (
	"internal/itoa"
	"unsafe"
)

// Flags stored in the first byte of the struct field byte array. Must be kept
// up to date with compiler/interface.go.
const (
	structFieldFlagAnonymous = 1 << iota
	structFieldFlagHasTag
	structFieldFlagIsExported
	structFieldFlagIsEmbedded
)

type Kind uint8

// Copied from reflect/type.go
// https://golang.org/src/reflect/type.go?s=8302:8316#L217
// These constants must match basicTypes and the typeKind* constants in
// compiler/interface.go
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
	Pointer
	Slice
	Array
	Func
	Map
	Struct
)

// Ptr is the old name for the Pointer kind.
const Ptr = Pointer

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
	case Pointer:
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

// Copied from reflect/type.go
// https://go.dev/src/reflect/type.go?#L348

// ChanDir represents a channel type's direction.
type ChanDir int

const (
	RecvDir ChanDir             = 1 << iota // <-chan
	SendDir                                 // chan<-
	BothDir = RecvDir | SendDir             // chan
)

// Method represents a single method.
type Method struct {
	// Name is the method name.
	Name string

	// PkgPath is the package path that qualifies a lower case (unexported)
	// method name. It is empty for upper case (exported) method names.
	// The combination of PkgPath and Name uniquely identifies a method
	// in a method set.
	// See https://golang.org/ref/spec#Uniqueness_of_identifiers
	PkgPath string

	Type  Type  // method type
	Func  Value // func with receiver as first argument
	Index int   // index for Type.Method
}

// The following Type type has been copied almost entirely from
// https://github.com/golang/go/blob/go1.15/src/reflect/type.go#L27-L212.
// Some methods have been commented out as they haven't yet been implemented.

// Type is the representation of a Go type.
//
// Not all methods apply to all kinds of types. Restrictions,
// if any, are noted in the documentation for each method.
// Use the Kind method to find out the kind of type before
// calling kind-specific methods. Calling a method
// inappropriate to the kind of type causes a run-time panic.
//
// Type values are comparable, such as with the == operator,
// so they can be used as map keys.
// Two Type values are equal if they represent identical types.
type Type interface {
	// Methods applicable to all types.

	// Align returns the alignment in bytes of a value of
	// this type when allocated in memory.
	Align() int

	// FieldAlign returns the alignment in bytes of a value of
	// this type when used as a field in a struct.
	FieldAlign() int

	// Method returns the i'th method in the type's method set.
	// It panics if i is not in the range [0, NumMethod()).
	//
	// For a non-interface type T or *T, the returned Method's Type and Func
	// fields describe a function whose first argument is the receiver.
	//
	// For an interface type, the returned Method's Type field gives the
	// method signature, without a receiver, and the Func field is nil.
	//
	// Only exported methods are accessible and they are sorted in
	// lexicographic order.
	Method(int) Method

	// MethodByName returns the method with that name in the type's
	// method set and a boolean indicating if the method was found.
	//
	// For a non-interface type T or *T, the returned Method's Type and Func
	// fields describe a function whose first argument is the receiver.
	//
	// For an interface type, the returned Method's Type field gives the
	// method signature, without a receiver, and the Func field is nil.
	MethodByName(string) (Method, bool)

	// NumMethod returns the number of exported methods in the type's method set.
	NumMethod() int

	// Name returns the type's name within its package for a defined type.
	// For other (non-defined) types it returns the empty string.
	Name() string

	// PkgPath returns a defined type's package path, that is, the import path
	// that uniquely identifies the package, such as "encoding/base64".
	// If the type was predeclared (string, error) or not defined (*T, struct{},
	// []int, or A where A is an alias for a non-defined type), the package path
	// will be the empty string.
	PkgPath() string

	// Size returns the number of bytes needed to store
	// a value of the given type; it is analogous to unsafe.Sizeof.
	Size() uintptr

	// String returns a string representation of the type.
	// The string representation may use shortened package names
	// (e.g., base64 instead of "encoding/base64") and is not
	// guaranteed to be unique among types. To test for type identity,
	// compare the Types directly.
	String() string

	// Kind returns the specific kind of this type.
	Kind() Kind

	// Implements reports whether the type implements the interface type u.
	Implements(u Type) bool

	// AssignableTo reports whether a value of the type is assignable to type u.
	AssignableTo(u Type) bool

	// ConvertibleTo reports whether a value of the type is convertible to type u.
	ConvertibleTo(u Type) bool

	// Comparable reports whether values of this type are comparable.
	Comparable() bool

	// Methods applicable only to some types, depending on Kind.
	// The methods allowed for each kind are:
	//
	//	Int*, Uint*, Float*, Complex*: Bits
	//	Array: Elem, Len
	//	Chan: ChanDir, Elem
	//	Func: In, NumIn, Out, NumOut, IsVariadic.
	//	Map: Key, Elem
	//	Pointer: Elem
	//	Slice: Elem
	//	Struct: Field, FieldByIndex, FieldByName, FieldByNameFunc, NumField

	// Bits returns the size of the type in bits.
	// It panics if the type's Kind is not one of the
	// sized or unsized Int, Uint, Float, or Complex kinds.
	Bits() int

	// ChanDir returns a channel type's direction.
	// It panics if the type's Kind is not Chan.
	ChanDir() ChanDir

	// IsVariadic reports whether a function type's final input parameter
	// is a "..." parameter. If so, t.In(t.NumIn() - 1) returns the parameter's
	// implicit actual type []T.
	//
	// For concreteness, if t represents func(x int, y ... float64), then
	//
	//	t.NumIn() == 2
	//	t.In(0) is the reflect.Type for "int"
	//	t.In(1) is the reflect.Type for "[]float64"
	//	t.IsVariadic() == true
	//
	// IsVariadic panics if the type's Kind is not Func.
	IsVariadic() bool

	// Elem returns a type's element type.
	// It panics if the type's Kind is not Array, Chan, Map, Pointer, or Slice.
	Elem() Type

	// Field returns a struct type's i'th field.
	// It panics if the type's Kind is not Struct.
	// It panics if i is not in the range [0, NumField()).
	Field(i int) StructField

	// FieldByIndex returns the nested field corresponding
	// to the index sequence. It is equivalent to calling Field
	// successively for each index i.
	// It panics if the type's Kind is not Struct.
	FieldByIndex(index []int) StructField

	// FieldByName returns the struct field with the given name
	// and a boolean indicating if the field was found.
	FieldByName(name string) (StructField, bool)

	// FieldByNameFunc returns the struct field with a name
	// that satisfies the match function and a boolean indicating if
	// the field was found.
	//
	// FieldByNameFunc considers the fields in the struct itself
	// and then the fields in any embedded structs, in breadth first order,
	// stopping at the shallowest nesting depth containing one or more
	// fields satisfying the match function. If multiple fields at that depth
	// satisfy the match function, they cancel each other
	// and FieldByNameFunc returns no match.
	// This behavior mirrors Go's handling of name lookup in
	// structs containing embedded fields.
	//FieldByNameFunc(match func(string) bool) (StructField, bool)

	// In returns the type of a function type's i'th input parameter.
	// It panics if the type's Kind is not Func.
	// It panics if i is not in the range [0, NumIn()).
	In(i int) Type

	// Key returns a map type's key type.
	// It panics if the type's Kind is not Map.
	Key() Type

	// Len returns an array type's length.
	// It panics if the type's Kind is not Array.
	Len() int

	// NumField returns a struct type's field count.
	// It panics if the type's Kind is not Struct.
	NumField() int

	// NumIn returns a function type's input parameter count.
	// It panics if the type's Kind is not Func.
	NumIn() int

	// NumOut returns a function type's output parameter count.
	// It panics if the type's Kind is not Func.
	NumOut() int

	// Out returns the type of a function type's i'th output parameter.
	// It panics if the type's Kind is not Func.
	// It panics if i is not in the range [0, NumOut()).
	Out(i int) Type
}

// Constants for the 'meta' byte.
const (
	kindMask  = 31 // mask to apply to the meta byte to get the Kind value
	flagNamed = 32 // flag that is set if this is a named type
)

// The base type struct. All type structs start with this.
type rawType struct {
	meta uint8 // metadata byte, contains kind and flags (see contants above)
}

// All types that have an element type: named, chan, slice, array, map (but not
// pointer because it doesn't have ptrTo).
type elemType struct {
	rawType
	ptrTo *rawType
	elem  *rawType
}

type ptrType struct {
	rawType
	elem *rawType
}

type arrayType struct {
	rawType
	ptrTo    *rawType
	elem     *rawType
	arrayLen uintptr
}

type mapType struct {
	rawType
	ptrTo *rawType
	elem  *rawType
	key   *rawType
}

type namedType struct {
	rawType
	ptrTo *rawType
	elem  *rawType
	pkg   *byte
	name  [1]byte
}

// Type for struct types. The numField value is intentionally put before ptrTo
// for better struct packing on 32-bit and 64-bit architectures. On these
// architectures, the ptrTo field still has the same offset as in all the other
// type structs.
// The fields array isn't necessarily 1 structField long, instead it is as long
// as numFields. The array is given a length of 1 to satisfy the Go type
// checker.
type structType struct {
	rawType
	numField uint16
	ptrTo    *rawType
	pkgpath  *byte
	fields   [1]structField // the remaining fields are all of type structField
}

type structField struct {
	fieldType *rawType
	data      unsafe.Pointer // various bits of information, packed in a byte array
}

// Equivalent to (go/types.Type).Underlying(): if this is a named type return
// the underlying type, else just return the type itself.
func (t *rawType) underlying() *rawType {
	if t.isNamed() {
		return (*elemType)(unsafe.Pointer(t)).elem
	}
	return t
}

func (t *rawType) isNamed() bool {
	return t.meta&flagNamed != 0
}

func TypeOf(i interface{}) Type {
	if i == nil {
		return nil
	}
	typecode, _ := decomposeInterface(i)
	return (*rawType)(typecode)
}

func PtrTo(t Type) Type { return PointerTo(t) }

func PointerTo(t Type) Type {
	return pointerTo(t.(*rawType))
}

func pointerTo(t *rawType) *rawType {
	switch t.Kind() {
	case Pointer:
		// TODO(dgryski): This is blocking https://github.com/tinygo-org/tinygo/issues/3131
		// We need to be able to create types that match existing types to prevent typecode equality.
		panic("reflect: cannot make **T type")
	case Struct:
		return (*structType)(unsafe.Pointer(t)).ptrTo
	default:
		return (*elemType)(unsafe.Pointer(t)).ptrTo
	}
}

func (t *rawType) String() string {
	if t.isNamed() {
		return t.PkgPath() + "." + t.Name()
	}

	switch t.Kind() {
	case Chan:
		return "chan " + t.elem().String()
	case Pointer:
		return "*" + t.elem().String()
	case Slice:
		return "[]" + t.elem().String()
	case Array:
		return "[" + itoa.Itoa(t.Len()) + "]" + t.elem().String()
	case Map:
		return "map[" + t.key().String() + "]" + t.elem().String()
	case Struct:
		s := "struct {"
		numField := t.NumField()
		for i := 0; i < numField; i++ {
			f := t.rawField(i)
			s += " " + f.Name + " " + f.Type.String()
			// every field except the last needs a semicolon
			if i < numField-1 {
				s += ";"
			}
		}
		s += " }"
		return s
	default:
		return t.Kind().String()
	}

	return t.Kind().String()
}

func (t *rawType) Kind() Kind {
	if t == nil {
		return Invalid
	}
	return Kind(t.meta & kindMask)
}

// Elem returns the element type for channel, slice and array types, the
// pointed-to value for pointer types, and the key type for map types.
func (t *rawType) Elem() Type {
	return t.elem()
}

func (t *rawType) elem() *rawType {
	underlying := t.underlying()
	switch underlying.Kind() {
	case Pointer:
		return (*ptrType)(unsafe.Pointer(underlying)).elem
	case Chan, Slice, Array, Map:
		return (*elemType)(unsafe.Pointer(underlying)).elem
	default:
		panic(&TypeError{"Elem"})
	}
}

func (t *rawType) key() *rawType {
	underlying := t.underlying()
	if underlying.Kind() != Map {
		panic(&TypeError{"Key"})
	}
	return (*mapType)(unsafe.Pointer(underlying)).key
}

// Field returns the type of the i'th field of this struct type. It panics if t
// is not a struct type.
func (t *rawType) Field(i int) StructField {
	field := t.rawField(i)
	return StructField{
		Name:      field.Name,
		PkgPath:   field.PkgPath,
		Type:      field.Type, // note: converts rawType to Type
		Tag:       field.Tag,
		Anonymous: field.Anonymous,
		Offset:    field.Offset,
		Index:     []int{i},
	}
}

func rawStructFieldFromPointer(descriptor *structType, fieldType *rawType, data unsafe.Pointer, flagsByte uint8, name string, offset uintptr) rawStructField {
	// Read the field tag, if there is one.
	var tag string
	if flagsByte&structFieldFlagHasTag != 0 {
		data = unsafe.Add(data, 1) // C: data+1
		tagLen := uintptr(*(*byte)(data))
		data = unsafe.Add(data, 1) // C: data+1
		tag = *(*string)(unsafe.Pointer(&stringHeader{
			data: data,
			len:  tagLen,
		}))
	}

	// Set the PkgPath to some (arbitrary) value if the package path is not
	// exported.
	pkgPath := ""
	if flagsByte&structFieldFlagIsExported == 0 {
		// This field is unexported.
		pkgPath = readStringZ(unsafe.Pointer(descriptor.pkgpath))
	}

	return rawStructField{
		Name:      name,
		PkgPath:   pkgPath,
		Type:      fieldType,
		Tag:       StructTag(tag),
		Anonymous: flagsByte&structFieldFlagAnonymous != 0,
		Offset:    offset,
	}
}

// rawField returns nearly the same value as Field but without converting the
// Type member to an interface.
//
// For internal use only.
func (t *rawType) rawField(n int) rawStructField {
	if t.Kind() != Struct {
		panic(&TypeError{"Field"})
	}
	descriptor := (*structType)(unsafe.Pointer(t.underlying()))
	if uint(n) >= uint(descriptor.numField) {
		panic("reflect: field index out of range")
	}

	// Iterate over all the fields to calculate the offset.
	// This offset could have been stored directly in the array (to make the
	// lookup faster), but by calculating it on-the-fly a bit of storage can be
	// saved.
	field := &descriptor.fields[0]
	var offset uintptr = 0
	for i := 0; i < n; i++ {
		offset += field.fieldType.Size()

		// Increment pointer to the next field.
		field = (*structField)(unsafe.Add(unsafe.Pointer(field), unsafe.Sizeof(structField{})))

		// Align the offset for the next field.
		offset = align(offset, uintptr(field.fieldType.Align()))
	}

	data := field.data

	// Read some flags of this field, like whether the field is an embedded
	// field. See structFieldFlagAnonymous and similar flags.
	flagsByte := *(*byte)(data)
	data = unsafe.Add(data, 1)

	name := readStringZ(data)
	data = unsafe.Add(data, len(name))

	return rawStructFieldFromPointer(descriptor, field.fieldType, data, flagsByte, name, offset)
}

// rawFieldByName returns nearly the same value as FieldByName but without converting the
// Type member to an interface.
//
// For internal use only.
func (t *rawType) rawFieldByName(n string) (rawStructField, []int, bool) {
	if t.Kind() != Struct {
		panic(&TypeError{"Field"})
	}

	type fieldWalker struct {
		t     *rawType
		index []int
	}

	queue := make([]fieldWalker, 0, 4)
	queue = append(queue, fieldWalker{t, nil})

	for len(queue) > 0 {
		type result struct {
			r     rawStructField
			index []int
		}

		var found []result
		var nextlevel []fieldWalker

		// For all the structs at this level..
		for _, ll := range queue {
			// Iterate over all the fields looking for the matching name
			// Also calculate field offset.

			descriptor := (*structType)(unsafe.Pointer(ll.t.underlying()))
			var offset uintptr
			field := &descriptor.fields[0]

			for i := uint16(0); i < descriptor.numField; i++ {
				data := field.data

				// Read some flags of this field, like whether the field is an embedded
				// field. See structFieldFlagAnonymous and similar flags.
				flagsByte := *(*byte)(data)
				data = unsafe.Add(data, 1)

				name := readStringZ(data)
				data = unsafe.Add(data, len(name))
				if name == n {
					found = append(found, result{
						rawStructFieldFromPointer(descriptor, field.fieldType, data, flagsByte, name, offset),
						append(ll.index, int(i)),
					})
				}

				structOrPtrToStruct := field.fieldType.Kind() == Struct || (field.fieldType.Kind() == Pointer && field.fieldType.elem().Kind() == Struct)
				if flagsByte&structFieldFlagIsEmbedded == structFieldFlagIsEmbedded && structOrPtrToStruct {
					embedded := field.fieldType
					if embedded.Kind() == Pointer {
						embedded = embedded.elem()
					}

					nextlevel = append(nextlevel, fieldWalker{
						t:     embedded,
						index: append(ll.index, int(i)),
					})
				}

				offset += field.fieldType.Size()

				// update offset/field pointer if there *is* a next field
				if i < descriptor.numField-1 {

					// Increment pointer to the next field.
					field = (*structField)(unsafe.Add(unsafe.Pointer(field), unsafe.Sizeof(structField{})))

					// Align the offset for the next field.
					offset = align(offset, uintptr(field.fieldType.Align()))
				}
			}
		}

		// found multiple hits at this level
		if len(found) > 1 {
			return rawStructField{}, nil, false
		}

		// found the field we were looking for
		if len(found) == 1 {
			r := found[0]
			return r.r, r.index, true
		}

		// else len(found) == 0, move on to the next level
		queue = append(queue[:0], nextlevel...)
	}

	// didn't find it
	return rawStructField{}, nil, false
}

// Bits returns the number of bits that this type uses. It is only valid for
// arithmetic types (integers, floats, and complex numbers). For other types, it
// will panic.
func (t *rawType) Bits() int {
	kind := t.Kind()
	if kind >= Int && kind <= Complex128 {
		return int(t.Size()) * 8
	}
	panic(TypeError{"Bits"})
}

// Len returns the number of elements in this array. It panics of the type kind
// is not Array.
func (t *rawType) Len() int {
	if t.Kind() != Array {
		panic(TypeError{"Len"})
	}

	return int((*arrayType)(unsafe.Pointer(t.underlying())).arrayLen)
}

// NumField returns the number of fields of a struct type. It panics for other
// type kinds.
func (t *rawType) NumField() int {
	if t.Kind() != Struct {
		panic(&TypeError{"NumField"})
	}
	return int((*structType)(unsafe.Pointer(t.underlying())).numField)
}

// Size returns the size in bytes of a given type. It is similar to
// unsafe.Sizeof.
func (t *rawType) Size() uintptr {
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
		return unsafe.Sizeof("")
	case UnsafePointer, Chan, Map, Pointer:
		return unsafe.Sizeof(uintptr(0))
	case Slice:
		return unsafe.Sizeof([]int{})
	case Interface:
		return unsafe.Sizeof(interface{}(nil))
	case Func:
		var f func()
		return unsafe.Sizeof(f)
	case Array:
		return t.elem().Size() * uintptr(t.Len())
	case Struct:
		numField := t.NumField()
		if numField == 0 {
			return 0
		}
		lastField := t.rawField(numField - 1)
		return align(lastField.Offset+lastField.Type.Size(), uintptr(t.Align()))
	default:
		panic("unimplemented: size of type")
	}
}

// Align returns the alignment of this type. It is similar to calling
// unsafe.Alignof.
func (t *rawType) Align() int {
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
		return int(unsafe.Alignof(""))
	case UnsafePointer, Chan, Map, Pointer:
		return int(unsafe.Alignof(uintptr(0)))
	case Slice:
		return int(unsafe.Alignof([]int(nil)))
	case Interface:
		return int(unsafe.Alignof(interface{}(nil)))
	case Func:
		var f func()
		return int(unsafe.Alignof(f))
	case Struct:
		numField := t.NumField()
		alignment := 1
		for i := 0; i < numField; i++ {
			fieldAlignment := t.rawField(i).Type.Align()
			if fieldAlignment > alignment {
				alignment = fieldAlignment
			}
		}
		return alignment
	case Array:
		return t.elem().Align()
	default:
		panic("unimplemented: alignment of type")
	}
}

// FieldAlign returns the alignment if this type is used in a struct field. It
// is currently an alias for Align() but this might change in the future.
func (t *rawType) FieldAlign() int {
	return t.Align()
}

// AssignableTo returns whether a value of type t can be assigned to a variable
// of type u.
func (t *rawType) AssignableTo(u Type) bool {
	if t == u.(*rawType) {
		return true
	}

	if u.Kind() == Interface && u.NumMethod() == 0 {
		return true
	}

	if u.Kind() == Interface {
		panic("reflect: unimplemented: AssignableTo with interface")
	}
	return false
}

func (t *rawType) Implements(u Type) bool {
	if u.Kind() != Interface {
		panic("reflect: non-interface type passed to Type.Implements")
	}
	return t.AssignableTo(u)
}

// Comparable returns whether values of this type can be compared to each other.
func (t *rawType) Comparable() bool {
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
	case Pointer:
		return true
	case Slice:
		return false
	case Array:
		return t.elem().Comparable()
	case Func:
		return false
	case Map:
		return false
	case Struct:
		numField := t.NumField()
		for i := 0; i < numField; i++ {
			if !t.rawField(i).Type.Comparable() {
				return false
			}
		}
		return true
	default:
		panic(TypeError{"Comparable"})
	}
}

// isbinary() returns if the hashmapAlgorithmBinary functions can be used on this type
func (t *rawType) isBinary() bool {
	switch t.Kind() {
	case Bool, Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
		return true
	case Pointer:
		return true
	case Array:
		return t.elem().isBinary()
	case Struct:
		numField := t.NumField()
		for i := 0; i < numField; i++ {
			if !t.rawField(i).Type.isBinary() {
				return false
			}
		}
		return true
	}
	return false
}

func (t rawType) ChanDir() ChanDir {
	panic("unimplemented: (reflect.Type).ChanDir()")
}

func (t *rawType) ConvertibleTo(u Type) bool {
	panic("unimplemented: (reflect.Type).ConvertibleTo()")
}

func (t *rawType) IsVariadic() bool {
	panic("unimplemented: (reflect.Type).IsVariadic()")
}

func (t *rawType) NumIn() int {
	panic("unimplemented: (reflect.Type).NumIn()")
}

func (t *rawType) NumOut() int {
	panic("unimplemented: (reflect.Type).NumOut()")
}

func (t *rawType) NumMethod() int {
	panic("unimplemented: (reflect.Type).NumMethod()")
}

// Read and return a null terminated string starting from data.
func readStringZ(data unsafe.Pointer) string {
	start := data
	var len uintptr
	for *(*byte)(data) != 0 {
		len++
		data = unsafe.Add(data, 1) // C: data++
	}

	return *(*string)(unsafe.Pointer(&stringHeader{
		data: start,
		len:  len,
	}))
}

func (t *rawType) Name() string {
	if t.isNamed() {
		ntype := (*namedType)(unsafe.Pointer(t))
		return readStringZ(unsafe.Pointer(&ntype.name[0]))
	}

	if t.Kind() <= UnsafePointer {
		return t.Kind().String()
	}

	return ""
}

func (t *rawType) Key() Type {
	return t.key()
}

func (t rawType) In(i int) Type {
	panic("unimplemented: (reflect.Type).In()")
}

func (t rawType) Out(i int) Type {
	panic("unimplemented: (reflect.Type).Out()")
}

func (t rawType) Method(i int) Method {
	panic("unimplemented: (reflect.Type).Method()")
}

func (t rawType) MethodByName(name string) (Method, bool) {
	panic("unimplemented: (reflect.Type).MethodByName()")
}

func (t *rawType) PkgPath() string {
	if t.isNamed() {
		ntype := (*namedType)(unsafe.Pointer(t))
		return readStringZ(unsafe.Pointer(ntype.pkg))
	}

	return ""
}

func (t *rawType) FieldByName(name string) (StructField, bool) {
	if t.Kind() != Struct {
		panic(TypeError{"FieldByName"})
	}

	field, index, ok := t.rawFieldByName(name)
	if !ok {
		return StructField{}, false
	}

	return StructField{
		Name:      field.Name,
		PkgPath:   field.PkgPath,
		Type:      field.Type, // note: converts rawType to Type
		Tag:       field.Tag,
		Anonymous: field.Anonymous,
		Offset:    field.Offset,
		Index:     index,
	}, true
}

func (t *rawType) FieldByIndex(index []int) StructField {
	ftype := t
	var field rawStructField

	for _, n := range index {
		structOrPtrToStruct := ftype.Kind() == Struct || (ftype.Kind() == Pointer && ftype.elem().Kind() == Struct)
		if !structOrPtrToStruct {
			panic(&TypeError{"FieldByIndex:" + ftype.Kind().String()})
		}

		if ftype.Kind() == Pointer {
			ftype = ftype.elem()
		}

		field = ftype.rawField(n)
		ftype = field.Type
	}

	return StructField{
		Name:      field.Name,
		PkgPath:   field.PkgPath,
		Type:      field.Type, // note: converts rawType to Type
		Tag:       field.Tag,
		Anonymous: field.Anonymous,
		Offset:    field.Offset,
		Index:     index,
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
	Tag       StructTag // field tag string
	Offset    uintptr
	Index     []int // index sequence for Type.FieldByIndex
	Anonymous bool
}

// IsExported reports whether the field is exported.
func (f StructField) IsExported() bool {
	return f.PkgPath == ""
}

// rawStructField is the same as StructField but with the Type member replaced
// with rawType. For internal use only. Avoiding this conversion to the Type
// interface improves code size in many cases.
type rawStructField struct {
	Name      string
	PkgPath   string
	Type      *rawType
	Tag       StructTag
	Offset    uintptr
	Anonymous bool
}

// A StructTag is the tag string in a struct field.
type StructTag string

// TODO: it would be feasible to do the key/value splitting at compile time,
// avoiding the code size cost of doing it at runtime

// Get returns the value associated with key in the tag string.
func (tag StructTag) Get(key string) string {
	v, _ := tag.Lookup(key)
	return v
}

// Lookup returns the value associated with key in the tag string.
func (tag StructTag) Lookup(key string) (value string, ok bool) {
	for tag != "" {
		// Skip leading space.
		i := 0
		for i < len(tag) && tag[i] == ' ' {
			i++
		}
		tag = tag[i:]
		if tag == "" {
			break
		}

		// Scan to colon. A space, a quote or a control character is a syntax error.
		// Strictly speaking, control chars include the range [0x7f, 0x9f], not just
		// [0x00, 0x1f], but in practice, we ignore the multi-byte control characters
		// as it is simpler to inspect the tag's bytes than the tag's runes.
		i = 0
		for i < len(tag) && tag[i] > ' ' && tag[i] != ':' && tag[i] != '"' && tag[i] != 0x7f {
			i++
		}
		if i == 0 || i+1 >= len(tag) || tag[i] != ':' || tag[i+1] != '"' {
			break
		}
		name := string(tag[:i])
		tag = tag[i+1:]

		// Scan quoted string to find value.
		i = 1
		for i < len(tag) && tag[i] != '"' {
			if tag[i] == '\\' {
				i++
			}
			i++
		}
		if i >= len(tag) {
			break
		}
		qvalue := string(tag[:i+1])
		tag = tag[i+1:]

		if key == name {
			value, err := unquote(qvalue)
			if err != nil {
				break
			}
			return value, true
		}
	}
	return "", false
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

func SliceOf(t Type) Type {
	panic("unimplemented: reflect.SliceOf()")
}

func ArrayOf(n int, t Type) Type {
	panic("unimplemented: reflect.ArrayOf()")
}

func StructOf([]StructField) Type {
	panic("unimplemented: reflect.StructOf()")
}

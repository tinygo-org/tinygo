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
// - channels and slices (see elemType):
//     meta          uint8
//     nmethods     uint16 (0)
//     ptrTo        *typeStruct
//     elementType  *typeStruct // the type that you get with .Elem()
// - pointer types (see ptrType, this doesn't include chan, map, etc):
//     meta         uint8
//     nmethods     uint16
//     elementType  *typeStruct
// - array types (see arrayType)
//     meta         uint8
//     nmethods     uint16 (0)
//     ptrTo        *typeStruct
//     elem         *typeStruct // element type of the array
//     arrayLen     uintptr     // length of the array (this is part of the type)
//     slicePtr     *typeStruct // pointer to []T type
// - map types (this is still missing the key and element types)
//     meta         uint8
//     nmethods     uint16 (0)
//     ptrTo        *typeStruct
//     elem         *typeStruct
//     key          *typeStruct
// - struct types (see structType):
//     meta         uint8
//     nmethods     uint16
//     ptrTo        *typeStruct
//     size         uint32
//     pkgpath      *byte       // package path; null terminated
//     numField     uint16
//     fields       [...]structField // the remaining fields are all of type structField
// - interface types (this is missing the interface methods):
//     meta         uint8
//     ptrTo        *typeStruct
// - signature types (this is missing input and output parameters):
//     meta         uint8
//     ptrTo        *typeStruct
// - named types
//     meta         uint8
//     nmethods     uint16      // number of methods
//     ptrTo        *typeStruct
//     elem         *typeStruct // underlying type
//     pkgpath      *byte       // pkgpath; null terminated
//     name         [1]byte     // actual name; null terminated
//
// The type struct is essentially a union of all the above types. Which it is,
// can be determined by looking at the meta byte.

package reflect

import (
	"internal/reflectlite"
	"unsafe"
)

type Kind = reflectlite.Kind

const (
	Invalid       Kind = reflectlite.Invalid
	Bool          Kind = reflectlite.Bool
	Int           Kind = reflectlite.Int
	Int8          Kind = reflectlite.Int8
	Int16         Kind = reflectlite.Int16
	Int32         Kind = reflectlite.Int32
	Int64         Kind = reflectlite.Int64
	Uint          Kind = reflectlite.Uint
	Uint8         Kind = reflectlite.Uint8
	Uint16        Kind = reflectlite.Uint16
	Uint32        Kind = reflectlite.Uint32
	Uint64        Kind = reflectlite.Uint64
	Uintptr       Kind = reflectlite.Uintptr
	Float32       Kind = reflectlite.Float32
	Float64       Kind = reflectlite.Float64
	Complex64     Kind = reflectlite.Complex64
	Complex128    Kind = reflectlite.Complex128
	Array         Kind = reflectlite.Array
	Chan          Kind = reflectlite.Chan
	Func          Kind = reflectlite.Func
	Interface     Kind = reflectlite.Interface
	Map           Kind = reflectlite.Map
	Pointer       Kind = reflectlite.Pointer
	Slice         Kind = reflectlite.Slice
	String        Kind = reflectlite.String
	Struct        Kind = reflectlite.Struct
	UnsafePointer Kind = reflectlite.UnsafePointer
)

const Ptr = reflectlite.Ptr

type ChanDir = reflectlite.ChanDir

const (
	RecvDir = reflectlite.RecvDir
	SendDir = reflectlite.SendDir
	BothDir = reflectlite.BothDir
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

// IsExported reports whether the method is exported.
func (m Method) IsExported() bool {
	return m.PkgPath == ""
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
	FieldByNameFunc(match func(string) bool) (StructField, bool)

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

	// OverflowComplex reports whether the complex128 x cannot be represented by type t.
	// It panics if t's Kind is not Complex64 or Complex128.
	OverflowComplex(x complex128) bool

	// OverflowFloat reports whether the float64 x cannot be represented by type t.
	// It panics if t's Kind is not Float32 or Float64.
	OverflowFloat(x float64) bool

	// OverflowInt reports whether the int64 x cannot be represented by type t.
	// It panics if t's Kind is not Int, Int8, Int16, Int32, or Int64.
	OverflowInt(x int64) bool

	// OverflowUint reports whether the uint64 x cannot be represented by type t.
	// It panics if t's Kind is not Uint, Uintptr, Uint8, Uint16, Uint32, or Uint64.
	OverflowUint(x uint64) bool

	// CanSeq reports whether a [Value] with this type can be iterated over using [Value.Seq].
	CanSeq() bool

	// CanSeq2 reports whether a [Value] with this type can be iterated over using [Value.Seq2].
	CanSeq2() bool
}

type rawType struct {
	reflectlite.RawType
}

func toType(t reflectlite.Type) Type {
	if t == nil {
		return nil
	}
	return (*rawType)(unsafe.Pointer(t.(*reflectlite.RawType)))
}

func toRawType(t Type) *reflectlite.RawType {
	return (*reflectlite.RawType)(unsafe.Pointer(t.(*rawType)))
}

func TypeOf(i interface{}) Type {
	return toType(reflectlite.TypeOf(i))
}

func PtrTo(t Type) Type {
	return PointerTo(t)
}

func PointerTo(t Type) Type {
	return toType(reflectlite.PointerTo(toRawType(t)))
}

func (t *rawType) AssignableTo(u Type) bool {
	return t.RawType.AssignableTo(&(u.(*rawType).RawType))
}

func (t *rawType) CanSeq() bool {
	switch t.Kind() {
	case Int8, Int16, Int32, Int64, Int, Uint8, Uint16, Uint32, Uint64, Uint, Uintptr, Array, Slice, Chan, String, Map:
		return true
	case Func:
		// TODO: implement canRangeFunc
		// return canRangeFunc(t)
		panic("unimplemented: (reflect.Type).CanSeq() for functions")
	case Pointer:
		return t.Elem().Kind() == Array
	}
	return false
}

func (t *rawType) CanSeq2() bool {
	switch t.Kind() {
	case Array, Slice, String, Map:
		return true
	case Func:
		// TODO: implement canRangeFunc2
		// return canRangeFunc2(t)
		panic("unimplemented: (reflect.Type).CanSeq2() for functions")
	case Pointer:
		return t.Elem().Kind() == Array
	}
	return false
}

func (t *rawType) ConvertibleTo(u Type) bool {
	panic("unimplemented: (reflect.Type).ConvertibleTo()")
}

func (t *rawType) Elem() Type {
	return toType(t.RawType.Elem())
}

func (t *rawType) Field(i int) StructField {
	f := t.RawType.Field(i)
	return toStructField(f)
}

func (t *rawType) FieldByIndex(index []int) StructField {
	f := t.RawType.FieldByIndex(index)
	return toStructField(f)
}

func (t *rawType) FieldByName(name string) (StructField, bool) {
	f, ok := t.RawType.FieldByName(name)
	return toStructField(f), ok
}

func (t *rawType) FieldByNameFunc(match func(string) bool) (StructField, bool) {
	f, ok := t.RawType.FieldByNameFunc(match)
	return toStructField(f), ok
}

func (t *rawType) Implements(u Type) bool {
	return t.RawType.Implements(&(u.(*rawType).RawType))
}

func (t *rawType) In(i int) Type {
	panic("unimplemented: (reflect.Type).In()")
}

func (t *rawType) IsVariadic() bool {
	panic("unimplemented: (reflect.Type).IsVariadic()")
}

func (t *rawType) Key() Type {
	return toType(t.RawType.Key())
}

func (t *rawType) Method(i int) Method {
	panic("unimplemented: (reflect.Type).Method()")
}

func (t *rawType) MethodByName(name string) (Method, bool) {
	panic("unimplemented: (reflect.Type).MethodByName()")
}

func (t *rawType) NumIn() int {
	panic("unimplemented: (reflect.Type).NumIn()")
}

func (t *rawType) NumOut() int {
	panic("unimplemented: (reflect.Type).NumOut()")
}

func (t *rawType) Out(i int) Type {
	panic("unimplemented: (reflect.Type).Out()")
}

// A StructField describes a single field in a struct.
// This must be kept in sync with [reflectlite.StructField].
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

func toStructField(f reflectlite.StructField) StructField {
	return StructField{
		Name:      f.Name,
		PkgPath:   f.PkgPath,
		Type:      toType(f.Type),
		Tag:       f.Tag,
		Offset:    f.Offset,
		Index:     f.Index,
		Anonymous: f.Anonymous,
	}
}

// IsExported reports whether the field is exported.
func (f StructField) IsExported() bool {
	return f.PkgPath == ""
}

type StructTag = reflectlite.StructTag

func TypeFor[T any]() Type {
	return toType(reflectlite.TypeFor[T]())
}

func SliceOf(t Type) Type {
	return toType(reflectlite.SliceOf(toRawType(t)))
}

func ArrayOf(n int, t Type) Type {
	return toType(reflectlite.ArrayOf(n, toRawType(t)))
}

func StructOf([]StructField) Type {
	return toType(reflectlite.StructOf([]reflectlite.StructField{}))
}

func MapOf(key, value Type) Type {
	return toType(reflectlite.MapOf(toRawType(key), toRawType(value)))
}

func FuncOf(in, out []Type, variadic bool) Type {
	rawIn := make([]reflectlite.Type, len(in))
	for i, t := range in {
		rawIn[i] = toRawType(t)
	}

	rawOut := make([]reflectlite.Type, len(out))
	for i, t := range out {
		rawOut[i] = toRawType(t)
	}

	return toType(reflectlite.FuncOf(rawIn, rawOut, variadic))
}

func ChanOf(dir ChanDir, t Type) Type {
	panic("unimplemented: reflect.ChanOf")
}

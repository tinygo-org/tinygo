package reflectlite

import "reflect"

func Swapper(slice interface{}) func(i, j int) {
	return reflect.Swapper(slice)
}

type Kind = reflect.Kind
type Type = reflect.Type
type Value = reflect.Value

const (
	Invalid       Kind = reflect.Invalid
	Bool          Kind = reflect.Bool
	Int           Kind = reflect.Int
	Int8          Kind = reflect.Int8
	Int16         Kind = reflect.Int16
	Int32         Kind = reflect.Int32
	Int64         Kind = reflect.Int64
	Uint          Kind = reflect.Uint
	Uint8         Kind = reflect.Uint8
	Uint16        Kind = reflect.Uint16
	Uint32        Kind = reflect.Uint32
	Uint64        Kind = reflect.Uint64
	Uintptr       Kind = reflect.Uintptr
	Float32       Kind = reflect.Float32
	Float64       Kind = reflect.Float64
	Complex64     Kind = reflect.Complex64
	Complex128    Kind = reflect.Complex128
	Array         Kind = reflect.Array
	Chan          Kind = reflect.Chan
	Func          Kind = reflect.Func
	Interface     Kind = reflect.Interface
	Map           Kind = reflect.Map
	Ptr           Kind = reflect.Ptr
	Slice         Kind = reflect.Slice
	String        Kind = reflect.String
	Struct        Kind = reflect.Struct
	UnsafePointer Kind = reflect.UnsafePointer
)

func ValueOf(i interface{}) reflect.Value {
	return reflect.ValueOf(i)
}

func TypeOf(i interface{}) reflect.Type {
	return reflect.TypeOf(i)
}

type ValueError = reflect.ValueError

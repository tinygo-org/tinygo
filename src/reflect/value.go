package reflect

import (
	"unsafe"
)

// This is the same thing as an interface{}.
type Value struct {
	typecode Type
	value    unsafe.Pointer
}

func Indirect(v Value) Value {
	return v
}

func ValueOf(i interface{}) Value {
	return *(*Value)(unsafe.Pointer(&i))
}

func (v Value) Interface() interface{}

func (v Value) Type() Type {
	return v.typecode
}

func (v Value) Kind() Kind {
	return Invalid // TODO
}

func (v Value) IsNil() bool {
	panic("unimplemented: (reflect.Value).IsNil()")
}

func (v Value) Pointer() uintptr {
	panic("unimplemented: (reflect.Value).Pointer()")
}

func (v Value) IsValid() bool {
	panic("unimplemented: (reflect.Value).IsValid()")
}

func (v Value) CanInterface() bool {
	panic("unimplemented: (reflect.Value).CanInterface()")
}

func (v Value) CanAddr() bool {
	panic("unimplemented: (reflect.Value).CanAddr()")
}

func (v Value) Addr() Value {
	panic("unimplemented: (reflect.Value).Addr()")
}

func (v Value) CanSet() bool {
	panic("unimplemented: (reflect.Value).CanSet()")
}

func (v Value) Bool() bool {
	panic("unimplemented: (reflect.Value).Bool()")
}

func (v Value) Int() int64 {
	panic("unimplemented: (reflect.Value).Int()")
}

func (v Value) Uint() uint64 {
	panic("unimplemented: (reflect.Value).Uint()")
}

func (v Value) Float() float64 {
	panic("unimplemented: (reflect.Value).Float()")
}

func (v Value) Complex() complex128 {
	panic("unimplemented: (reflect.Value).Complex()")
}

func (v Value) String() string {
	panic("unimplemented: (reflect.Value).String()")
}

func (v Value) Bytes() []byte {
	panic("unimplemented: (reflect.Value).Bytes()")
}

func (v Value) Slice(i, j int) Value {
	panic("unimplemented: (reflect.Value).Slice()")
}

func (v Value) Len() int {
	panic("unimplemented: (reflect.Value).Len()")
}

func (v Value) NumField() int {
	panic("unimplemented: (reflect.Value).NumField()")
}

func (v Value) Elem() Value {
	panic("unimplemented: (reflect.Value).Elem()")
}

func (v Value) Field(i int) Value {
	panic("unimplemented: (reflect.Value).Field()")
}

func (v Value) Index(i int) Value {
	panic("unimplemented: (reflect.Value).Index()")
}

func (v Value) MapKeys() []Value {
	panic("unimplemented: (reflect.Value).MapKeys()")
}

func (v Value) MapIndex(key Value) Value {
	panic("unimplemented: (reflect.Value).MapIndex()")
}

func (v Value) Set(x Value) {
	panic("unimplemented: (reflect.Value).Set()")
}

func (v Value) SetBool(x bool) {
	panic("unimplemented: (reflect.Value).SetBool()")
}

func (v Value) SetInt(x int64) {
	panic("unimplemented: (reflect.Value).SetInt()")
}

func (v Value) SetUint(x uint64) {
	panic("unimplemented: (reflect.Value).SetUint()")
}

func (v Value) SetFloat(x float64) {
	panic("unimplemented: (reflect.Value).SetFloat()")
}

func (v Value) SetComplex(x complex128) {
	panic("unimplemented: (reflect.Value).SetComplex()")
}

func (v Value) SetString(x string) {
	panic("unimplemented: (reflect.Value).SetString()")
}

func MakeSlice(typ Type, len, cap int) Value {
	panic("unimplemented: reflect.MakeSlice()")
}

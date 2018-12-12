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
	switch v.Type().Kind() {
	case UnsafePointer:
		return uintptr(v.value)
	case Chan, Func, Map, Ptr, Slice:
		panic("unimplemented: (reflect.Value).Pointer()")
	default:
		panic(&ValueError{"Pointer"})
	}
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
	switch v.Type().Kind() {
	case Bool:
		return uintptr(v.value) != 0
	default:
		panic(&ValueError{"Bool"})
	}
}

func (v Value) Int() int64 {
	switch v.Type().Kind() {
	case Int:
		if unsafe.Sizeof(int(0)) <= unsafe.Sizeof(uintptr(0)) {
			return int64(int(uintptr(v.value)))
		} else {
			return int64(*(*int)(v.value))
		}
	case Int8:
		return int64(int8(uintptr(v.value)))
	case Int16:
		return int64(int16(uintptr(v.value)))
	case Int32:
		if unsafe.Sizeof(int32(0)) <= unsafe.Sizeof(uintptr(0)) {
			return int64(int32(uintptr(v.value)))
		} else {
			return int64(*(*int32)(v.value))
		}
		return int64(uintptr(v.value))
	case Int64:
		if unsafe.Sizeof(int64(0)) <= unsafe.Sizeof(uintptr(0)) {
			return int64(uintptr(v.value))
		} else {
			return *(*int64)(v.value)
		}
	default:
		panic(&ValueError{"Int"})
	}
}

func (v Value) Uint() uint64 {
	switch v.Type().Kind() {
	case Uintptr, Uint8, Uint16:
		return uint64(uintptr(v.value))
	case Uint:
		if unsafe.Sizeof(uint(0)) <= unsafe.Sizeof(uintptr(0)) {
			return uint64(uintptr(v.value))
		} else {
			// For systems with 16-bit pointers.
			return uint64(*(*uint)(v.value))
		}
	case Uint32:
		if unsafe.Sizeof(uint32(0)) <= unsafe.Sizeof(uintptr(0)) {
			return uint64(uintptr(v.value))
		} else {
			// For systems with 16-bit pointers.
			return uint64(*(*uint32)(v.value))
		}
	case Uint64:
		if unsafe.Sizeof(uint64(0)) <= unsafe.Sizeof(uintptr(0)) {
			return uint64(uintptr(v.value))
		} else {
			// For systems with 16-bit or 32-bit pointers.
			return *(*uint64)(v.value)
		}
	default:
		panic(&ValueError{"Uint"})
	}
}

func (v Value) Float() float64 {
	switch v.Type().Kind() {
	case Float32:
		if unsafe.Sizeof(float32(0)) <= unsafe.Sizeof(uintptr(0)) {
			// The float is directly stored in the interface value on systems
			// with 32-bit and 64-bit pointers.
			return float64(*(*float32)(unsafe.Pointer(&v.value)))
		} else {
			// The float is stored as an external value on systems with 16-bit
			// pointers.
			return float64(*(*float32)(v.value))
		}
	case Float64:
		if unsafe.Sizeof(float64(0)) <= unsafe.Sizeof(uintptr(0)) {
			// The float is directly stored in the interface value on systems
			// with 64-bit pointers.
			return *(*float64)(unsafe.Pointer(&v.value))
		} else {
			// For systems with 16-bit and 32-bit pointers.
			return *(*float64)(v.value)
		}
	default:
		panic(&ValueError{"Float"})
	}
}

func (v Value) Complex() complex128 {
	switch v.Type().Kind() {
	case Complex64:
		if unsafe.Sizeof(complex64(0)) <= unsafe.Sizeof(uintptr(0)) {
			// The complex number is directly stored in the interface value on
			// systems with 64-bit pointers.
			return complex128(*(*complex64)(unsafe.Pointer(&v.value)))
		} else {
			// The complex number is stored as an external value on systems with
			// 16-bit and 32-bit pointers.
			return complex128(*(*complex64)(v.value))
		}
	case Complex128:
		// This is a 128-bit value, which is always stored as an external value.
		// It may be stored in the pointer directly on very uncommon
		// architectures with 128-bit pointers, however.
		return *(*complex128)(v.value)
	default:
		panic(&ValueError{"Complex"})
	}
}

func (v Value) String() string {
	switch v.Type().Kind() {
	case String:
		// A string value is always bigger than a pointer as it is made of a
		// pointer and a length.
		return *(*string)(v.value)
	default:
		// Special case because of the special treatment of .String() in Go.
		return "<T>"
	}
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

type ValueError struct {
	Method string
}

func (e *ValueError) Error() string {
	return "reflect: call of reflect.Value." + e.Method + " on invalid type"
}

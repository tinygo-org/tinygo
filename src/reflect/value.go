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

func (v Value) Interface() interface{} {
	return *(*interface{})(unsafe.Pointer(&v))
}

func (v Value) Type() Type {
	return v.typecode
}

func (v Value) Kind() Kind {
	return v.Type().Kind()
}

func (v Value) IsNil() bool {
	panic("unimplemented: (reflect.Value).IsNil()")
}

func (v Value) Pointer() uintptr {
	switch v.Kind() {
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
	// No Value types of private data can be constructed at the moment.
	return true
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
	switch v.Kind() {
	case Bool:
		return uintptr(v.value) != 0
	default:
		panic(&ValueError{"Bool"})
	}
}

func (v Value) Int() int64 {
	switch v.Kind() {
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
	switch v.Kind() {
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
	switch v.Kind() {
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
	switch v.Kind() {
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
	switch v.Kind() {
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
	t := v.Type()
	switch t.Kind() {
	case Slice:
		return int((*SliceHeader)(v.value).Len)
	case String:
		return int((*StringHeader)(v.value).Len)
	default: // Array, Chan, Map
		panic("unimplemented: (reflect.Value).Len()")
	}
}

func (v Value) Cap() int {
	t := v.Type()
	switch t.Kind() {
	case Slice:
		return int((*SliceHeader)(v.value).Cap)
	default: // Array, Chan
		panic("unimplemented: (reflect.Value).Cap()")
	}
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
	switch v.Kind() {
	case Slice:
		// Extract an element from the slice.
		slice := *(*SliceHeader)(v.value)
		if uint(i) >= uint(slice.Len) {
			panic("reflect: slice index out of range")
		}
		elem := Value{
			typecode: v.Type().Elem(),
		}
		addr := uintptr(slice.Data) + elem.Type().Size() * uintptr(i) // pointer to new value
		if elem.Type().Size() <= unsafe.Sizeof(uintptr(0)) {
			// Value fits inside interface value.
			// Make sure to copy it from the slice to the interface value.
			var value uintptr
			for j := elem.Type().Size(); j != 0; j-- {
				value = (value << 8) | uintptr(*(*uint8)(unsafe.Pointer(addr + j - 1)))
			}
			elem.value = unsafe.Pointer(value)
		} else {
			// Value doesn't fit in the interface.
			// Store a pointer to the element in the interface.
			elem.value = unsafe.Pointer(addr)
		}
		return elem
	case String:
		// Extract a character from a string.
		// A string is never stored directly in the interface, but always as a
		// pointer to the string value.
		s := *(*StringHeader)(v.value)
		if uint(i) >= uint(s.Len) {
			panic("reflect: string index out of range")
		}
		return Value{
			typecode: Uint8.basicType(),
			value: unsafe.Pointer(uintptr(*(*uint8)(unsafe.Pointer(s.Data + uintptr(i))))),
		}
	case Array:
		panic("unimplemented: (reflect.Value).Index()")
	default:
		panic(&ValueError{"Index"})
	}
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

type SliceHeader struct {
	Data uintptr
	Len  uintptr
	Cap  uintptr
}

type StringHeader struct {
	Data uintptr
	Len  uintptr
}

type ValueError struct {
	Method string
}

func (e *ValueError) Error() string {
	return "reflect: call of reflect.Value." + e.Method + " on invalid type"
}

package reflect

import (
	"unsafe"
)

type Value struct {
	typecode Type
	value    unsafe.Pointer
	indirect bool
}

func Indirect(v Value) Value {
	if v.Kind() != Ptr {
		return v
	}
	return v.Elem()
}

func ValueOf(i interface{}) Value {
	v := (*interfaceHeader)(unsafe.Pointer(&i))
	return Value{
		typecode: v.typecode,
		value:    v.value,
	}
}

func (v Value) Interface() interface{} {
	i := interfaceHeader{
		typecode: v.typecode,
		value:    v.value,
	}
	if v.indirect && v.Type().Size() <= unsafe.Sizeof(uintptr(0)) {
		// Value was indirect but must be put back directly in the interface
		// value.
		var value uintptr
		for j := v.Type().Size(); j != 0; j-- {
			value = (value << 8) | uintptr(*(*uint8)(unsafe.Pointer(uintptr(v.value) + j - 1)))
		}
		i.value = unsafe.Pointer(value)
	}
	return *(*interface{})(unsafe.Pointer(&i))
}

func (v Value) Type() Type {
	return v.typecode
}

func (v Value) Kind() Kind {
	return v.Type().Kind()
}

func (v Value) IsNil() bool {
	switch v.Kind() {
	case Chan, Map, Ptr:
		return v.value == nil
	case Func:
		if v.value == nil {
			return true
		}
		fn := (*funcHeader)(v.value)
		return fn.Code == nil
	case Slice:
		if v.value == nil {
			return true
		}
		slice := (*SliceHeader)(v.value)
		return slice.Data == 0
	case Interface:
		if v.value == nil {
			return true
		}
		itf := (*interfaceHeader)(v.value)
		return itf.value == nil
	default:
		panic(&ValueError{"IsNil"})
	}
}

func (v Value) Pointer() uintptr {
	switch v.Kind() {
	case Chan, Map, Ptr, UnsafePointer:
		return uintptr(v.value)
	case Slice:
		slice := (*SliceHeader)(v.value)
		return slice.Data
	case Func:
		panic("unimplemented: (reflect.Value).Pointer()")
	default: // not implemented: Func
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
	return v.indirect
}

func (v Value) Bool() bool {
	switch v.Kind() {
	case Bool:
		if v.indirect {
			return *((*bool)(v.value))
		} else {
			return uintptr(v.value) != 0
		}
	default:
		panic(&ValueError{"Bool"})
	}
}

func (v Value) Int() int64 {
	switch v.Kind() {
	case Int:
		if v.indirect || unsafe.Sizeof(int(0)) > unsafe.Sizeof(uintptr(0)) {
			return int64(*(*int)(v.value))
		} else {
			return int64(int(uintptr(v.value)))
		}
	case Int8:
		if v.indirect {
			return int64(*(*int8)(v.value))
		} else {
			return int64(int8(uintptr(v.value)))
		}
	case Int16:
		if v.indirect {
			return int64(*(*int16)(v.value))
		} else {
			return int64(int16(uintptr(v.value)))
		}
	case Int32:
		if v.indirect || unsafe.Sizeof(int32(0)) > unsafe.Sizeof(uintptr(0)) {
			return int64(*(*int32)(v.value))
		} else {
			return int64(int32(uintptr(v.value)))
		}
	case Int64:
		if v.indirect || unsafe.Sizeof(int64(0)) > unsafe.Sizeof(uintptr(0)) {
			return int64(*(*int64)(v.value))
		} else {
			return int64(int64(uintptr(v.value)))
		}
	default:
		panic(&ValueError{"Int"})
	}
}

func (v Value) Uint() uint64 {
	switch v.Kind() {
	case Uintptr:
		if v.indirect {
			return uint64(*(*uintptr)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint8:
		if v.indirect {
			return uint64(*(*uint8)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint16:
		if v.indirect {
			return uint64(*(*uint16)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint:
		if v.indirect || unsafe.Sizeof(uint(0)) > unsafe.Sizeof(uintptr(0)) {
			return uint64(*(*uint)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint32:
		if v.indirect || unsafe.Sizeof(uint32(0)) > unsafe.Sizeof(uintptr(0)) {
			return uint64(*(*uint32)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint64:
		if v.indirect || unsafe.Sizeof(uint64(0)) > unsafe.Sizeof(uintptr(0)) {
			return uint64(*(*uint64)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	default:
		panic(&ValueError{"Uint"})
	}
}

func (v Value) Float() float64 {
	switch v.Kind() {
	case Float32:
		if v.indirect || unsafe.Sizeof(float32(0)) > unsafe.Sizeof(uintptr(0)) {
			// The float is stored as an external value on systems with 16-bit
			// pointers.
			return float64(*(*float32)(v.value))
		} else {
			// The float is directly stored in the interface value on systems
			// with 32-bit and 64-bit pointers.
			return float64(*(*float32)(unsafe.Pointer(&v.value)))
		}
	case Float64:
		if v.indirect || unsafe.Sizeof(float64(0)) > unsafe.Sizeof(uintptr(0)) {
			// For systems with 16-bit and 32-bit pointers.
			return *(*float64)(v.value)
		} else {
			// The float is directly stored in the interface value on systems
			// with 64-bit pointers.
			return *(*float64)(unsafe.Pointer(&v.value))
		}
	default:
		panic(&ValueError{"Float"})
	}
}

func (v Value) Complex() complex128 {
	switch v.Kind() {
	case Complex64:
		if v.indirect || unsafe.Sizeof(complex64(0)) > unsafe.Sizeof(uintptr(0)) {
			// The complex number is stored as an external value on systems with
			// 16-bit and 32-bit pointers.
			return complex128(*(*complex64)(v.value))
		} else {
			// The complex number is directly stored in the interface value on
			// systems with 64-bit pointers.
			return complex128(*(*complex64)(unsafe.Pointer(&v.value)))
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
	switch v.Kind() {
	case Ptr:
		ptr := v.value
		if v.indirect {
			ptr = *(*unsafe.Pointer)(ptr)
		}
		if ptr == nil {
			return Value{}
		}
		return Value{
			typecode: v.Type().Elem(),
			value:    ptr,
			indirect: true,
		}
	default: // not implemented: Interface
		panic(&ValueError{"Elem"})
	}
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
			indirect: true,
		}
		addr := uintptr(slice.Data) + elem.Type().Size() * uintptr(i) // pointer to new value
		elem.value = unsafe.Pointer(addr)
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

func (v Value) MapRange() *MapIter {
	panic("unimplemented: (reflect.Value).MapRange()")
}

type MapIter struct {
}

func (it *MapIter) Key() Value {
	panic("unimplemented: (*reflect.MapIter).Key()")
}

func (it *MapIter) Value() Value {
	panic("unimplemented: (*reflect.MapIter).Value()")
}

func (it *MapIter) Next() bool {
	panic("unimplemented: (*reflect.MapIter).Next()")
}

func (v Value) Set(x Value) {
	if !v.indirect {
		panic("reflect: value is not addressable")
	}
	if v.Type() != x.Type() {
		if v.Kind() == Interface {
			panic("reflect: unimplemented: assigning to interface of different type")
		} else {
			panic("reflect: cannot assign")
		}
	}
	size := v.Type().Size()
	xptr := x.value
	if size <= unsafe.Sizeof(uintptr(0)) && !x.indirect {
		value := x.value
		xptr = unsafe.Pointer(&value)
	}
	memcpy(v.value, xptr, size)
}

func (v Value) SetBool(x bool) {
	if !v.indirect {
		panic("reflect: value is not addressable")
	}
	switch v.Kind() {
	case Bool:
		*(*bool)(v.value) = x
	default:
		panic(&ValueError{"SetBool"})
	}
}

func (v Value) SetInt(x int64) {
	if !v.indirect {
		panic("reflect: value is not addressable")
	}
	switch v.Kind() {
	case Int:
		*(*int)(v.value) = int(x)
	case Int8:
		*(*int8)(v.value) = int8(x)
	case Int16:
		*(*int16)(v.value) = int16(x)
	case Int32:
		*(*int32)(v.value) = int32(x)
	case Int64:
		*(*int64)(v.value) = x
	default:
		panic(&ValueError{"SetInt"})
	}
}

func (v Value) SetUint(x uint64) {
	if !v.indirect {
		panic("reflect: value is not addressable")
	}
	switch v.Kind() {
	case Uint:
		*(*uint)(v.value) = uint(x)
	case Uint8:
		*(*uint8)(v.value) = uint8(x)
	case Uint16:
		*(*uint16)(v.value) = uint16(x)
	case Uint32:
		*(*uint32)(v.value) = uint32(x)
	case Uint64:
		*(*uint64)(v.value) = x
	case Uintptr:
		*(*uintptr)(v.value) = uintptr(x)
	default:
		panic(&ValueError{"SetUint"})
	}
}

func (v Value) SetFloat(x float64) {
	if !v.indirect {
		panic("reflect: value is not addressable")
	}
	switch v.Kind() {
	case Float32:
		*(*float32)(v.value) = float32(x)
	case Float64:
		*(*float64)(v.value) = x
	default:
		panic(&ValueError{"SetFloat"})
	}
}

func (v Value) SetComplex(x complex128) {
	if !v.indirect {
		panic("reflect: value is not addressable")
	}
	switch v.Kind() {
	case Complex64:
		*(*complex64)(v.value) = complex64(x)
	case Complex128:
		*(*complex128)(v.value) = x
	default:
		panic(&ValueError{"SetComplex"})
	}
}

func (v Value) SetString(x string) {
	if !v.indirect {
		panic("reflect: value is not addressable")
	}
	switch v.Kind() {
	case String:
		*(*string)(v.value) = x
	default:
		panic(&ValueError{"SetString"})
	}
}

func MakeSlice(typ Type, len, cap int) Value {
	panic("unimplemented: reflect.MakeSlice()")
}

type funcHeader struct {
	Context unsafe.Pointer
	Code    unsafe.Pointer
}

// This is the same thing as an interface{}.
type interfaceHeader struct {
	typecode Type
	value    unsafe.Pointer
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

//go:linkname memcpy runtime.memcpy
func memcpy(dst, src unsafe.Pointer, size uintptr)

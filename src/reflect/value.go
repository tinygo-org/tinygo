package reflect

import (
	"math"
	"unsafe"
)

type valueFlags uint8

// Flags list some useful flags that contain some extra information not
// contained in an interface{} directly, like whether this value was exported at
// all (it is possible to read unexported fields using reflection, but it is not
// possible to modify them).
const (
	valueFlagIndirect valueFlags = 1 << iota
	valueFlagExported
	valueFlagEmbedRO
	valueFlagStickyRO

	valueFlagRO = valueFlagEmbedRO | valueFlagStickyRO
)

func (v valueFlags) ro() valueFlags {
	if v&valueFlagRO != 0 {
		return valueFlagStickyRO
	}
	return 0
}

type Value struct {
	typecode *rawType
	value    unsafe.Pointer
	flags    valueFlags
}

// isIndirect returns whether the value pointer in this Value is always a
// pointer to the value. If it is false, it is only a pointer to the value if
// the value is bigger than a pointer.
func (v Value) isIndirect() bool {
	return v.flags&valueFlagIndirect != 0
}

// isExported returns whether the value represented by this Value could be
// accessed without violating type system constraints. For example, it is not
// set for unexported struct fields.
func (v Value) isExported() bool {
	return v.flags&valueFlagExported != 0
}

func (v Value) isRO() bool {
	return v.flags&(valueFlagRO) != 0
}

func (v Value) checkRO() {
	if v.isRO() {
		panic("reflect: value is not settable")
	}
}

func Indirect(v Value) Value {
	if v.Kind() != Ptr {
		return v
	}
	return v.Elem()
}

//go:linkname composeInterface runtime.composeInterface
func composeInterface(unsafe.Pointer, unsafe.Pointer) interface{}

//go:linkname decomposeInterface runtime.decomposeInterface
func decomposeInterface(i interface{}) (unsafe.Pointer, unsafe.Pointer)

func ValueOf(i interface{}) Value {
	typecode, value := decomposeInterface(i)
	return Value{
		typecode: (*rawType)(typecode),
		value:    value,
		flags:    valueFlagExported,
	}
}

func (v Value) Interface() interface{} {
	if !v.isExported() {
		panic("(reflect.Value).Interface: unexported")
	}
	return valueInterfaceUnsafe(v)
}

// valueInterfaceUnsafe is used by the runtime to hash map keys. It should not
// be subject to the isExported check.
func valueInterfaceUnsafe(v Value) interface{} {
	if v.typecode.Kind() == Interface {
		// The value itself is an interface. This can happen when getting the
		// value of a struct field of interface type, like this:
		//     type T struct {
		//         X interface{}
		//     }
		return *(*interface{})(v.value)
	}
	if v.isIndirect() && v.typecode.Size() <= unsafe.Sizeof(uintptr(0)) {
		// Value was indirect but must be put back directly in the interface
		// value.
		var value uintptr
		for j := v.typecode.Size(); j != 0; j-- {
			value = (value << 8) | uintptr(*(*uint8)(unsafe.Add(v.value, j-1)))
		}
		v.value = unsafe.Pointer(value)
	}
	return composeInterface(unsafe.Pointer(v.typecode), v.value)
}

func (v Value) Type() Type {
	return v.typecode
}

// IsZero reports whether v is the zero value for its type.
// It panics if the argument is invalid.
func (v Value) IsZero() bool {
	switch v.Kind() {
	case Bool:
		return !v.Bool()
	case Int, Int8, Int16, Int32, Int64:
		return v.Int() == 0
	case Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
		return v.Uint() == 0
	case Float32, Float64:
		return math.Float64bits(v.Float()) == 0
	case Complex64, Complex128:
		c := v.Complex()
		return math.Float64bits(real(c)) == 0 && math.Float64bits(imag(c)) == 0
	case Array:
		for i := 0; i < v.Len(); i++ {
			if !v.Index(i).IsZero() {
				return false
			}
		}
		return true
	case Chan, Func, Interface, Map, Pointer, Slice, UnsafePointer:
		return v.IsNil()
	case String:
		return v.Len() == 0
	case Struct:
		for i := 0; i < v.NumField(); i++ {
			if !v.Field(i).IsZero() {
				return false
			}
		}
		return true
	default:
		// This should never happens, but will act as a safeguard for
		// later, as a default value doesn't makes sense here.
		panic(&ValueError{Method: "reflect.Value.IsZero", Kind: v.Kind()})
	}
}

// Internal function only, do not use.
//
// RawType returns the raw, underlying type code. It is used in the runtime
// package and needs to be exported for the runtime package to access it.
func (v Value) RawType() *rawType {
	return v.typecode
}

func (v Value) Kind() Kind {
	return v.typecode.Kind()
}

// IsNil returns whether the value is the nil value. It panics if the value Kind
// is not a channel, map, pointer, function, slice, or interface.
func (v Value) IsNil() bool {
	switch v.Kind() {
	case Chan, Map, Ptr, UnsafePointer:
		return v.pointer() == nil
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
		slice := (*sliceHeader)(v.value)
		return slice.data == nil
	case Interface:
		val := *(*interface{})(v.value)
		return val == nil
	default:
		panic(&ValueError{Method: "IsNil", Kind: v.Kind()})
	}
}

// Pointer returns the underlying pointer of the given value for the following
// types: chan, map, pointer, unsafe.Pointer, slice, func.
func (v Value) Pointer() uintptr {
	return uintptr(v.UnsafePointer())
}

// UnsafePointer returns the underlying pointer of the given value for the
// following types: chan, map, pointer, unsafe.Pointer, slice, func.
func (v Value) UnsafePointer() unsafe.Pointer {
	switch v.Kind() {
	case Chan, Map, Ptr, UnsafePointer:
		return v.pointer()
	case Slice:
		slice := (*sliceHeader)(v.value)
		return slice.data
	case Func:
		fn := (*funcHeader)(v.value)
		if fn.Context != nil {
			return fn.Context
		}
		return fn.Code
	default:
		panic(&ValueError{Method: "UnsafePointer", Kind: v.Kind()})
	}
}

// pointer returns the underlying pointer represented by v.
// v.Kind() must be Ptr, Map, Chan, or UnsafePointer
func (v Value) pointer() unsafe.Pointer {
	if v.isIndirect() {
		return *(*unsafe.Pointer)(v.value)
	}
	return v.value
}

func (v Value) IsValid() bool {
	return v.typecode != nil
}

func (v Value) CanInterface() bool {
	return v.isExported() && !v.isRO()
}

func (v Value) CanAddr() bool {
	return v.flags&(valueFlagIndirect) == valueFlagIndirect
}

func (v Value) Comparable() bool {
	k := v.Kind()
	switch k {
	case Invalid:
		return false

	case Array:
		switch v.Type().Elem().Kind() {
		case Interface, Array, Struct:
			for i := 0; i < v.Type().Len(); i++ {
				if !v.Index(i).Comparable() {
					return false
				}
			}
			return true
		}
		return v.Type().Comparable()

	case Interface:
		return v.Elem().Comparable()

	case Struct:
		for i := 0; i < v.NumField(); i++ {
			if !v.Field(i).Comparable() {
				return false
			}
		}
		return true

	default:
		return v.Type().Comparable()
	}
}

func (v Value) Addr() Value {
	if !v.CanAddr() {
		panic("reflect.Value.Addr of unaddressable value")
	}
	// Preserve flagRO instead of using v.flag.ro() so that
	// v.Addr().Elem() is equivalent to v (#32772)
	flags := v.flags & (valueFlagExported | valueFlagRO)
	return Value{
		typecode: pointerTo(v.typecode),
		value:    v.value,
		flags:    flags,
	}
}

func (v Value) UnsafeAddr() uintptr {
	return uintptr(v.Addr().UnsafePointer())
}

func (v Value) CanSet() bool {
	return v.flags&(valueFlagExported|valueFlagIndirect|valueFlagRO) == valueFlagExported|valueFlagIndirect
}

func (v Value) Bool() bool {
	switch v.Kind() {
	case Bool:
		if v.isIndirect() {
			return *((*bool)(v.value))
		} else {
			return uintptr(v.value) != 0
		}
	default:
		panic(&ValueError{Method: "Bool", Kind: v.Kind()})
	}
}

// CanInt reports whether Uint can be used without panicking.
func (v Value) CanInt() bool {
	switch v.Kind() {
	case Int, Int8, Int16, Int32, Int64:
		return true
	default:
		return false
	}
}

func (v Value) Int() int64 {
	switch v.Kind() {
	case Int:
		if v.isIndirect() || unsafe.Sizeof(int(0)) > unsafe.Sizeof(uintptr(0)) {
			return int64(*(*int)(v.value))
		} else {
			return int64(int(uintptr(v.value)))
		}
	case Int8:
		if v.isIndirect() {
			return int64(*(*int8)(v.value))
		} else {
			return int64(int8(uintptr(v.value)))
		}
	case Int16:
		if v.isIndirect() {
			return int64(*(*int16)(v.value))
		} else {
			return int64(int16(uintptr(v.value)))
		}
	case Int32:
		if v.isIndirect() || unsafe.Sizeof(int32(0)) > unsafe.Sizeof(uintptr(0)) {
			return int64(*(*int32)(v.value))
		} else {
			return int64(int32(uintptr(v.value)))
		}
	case Int64:
		if v.isIndirect() || unsafe.Sizeof(int64(0)) > unsafe.Sizeof(uintptr(0)) {
			return int64(*(*int64)(v.value))
		} else {
			return int64(int64(uintptr(v.value)))
		}
	default:
		panic(&ValueError{Method: "Int", Kind: v.Kind()})
	}
}

// CanUint reports whether Uint can be used without panicking.
func (v Value) CanUint() bool {
	switch v.Kind() {
	case Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
		return true
	default:
		return false
	}
}

func (v Value) Uint() uint64 {
	switch v.Kind() {
	case Uintptr:
		if v.isIndirect() {
			return uint64(*(*uintptr)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint8:
		if v.isIndirect() {
			return uint64(*(*uint8)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint16:
		if v.isIndirect() {
			return uint64(*(*uint16)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint:
		if v.isIndirect() || unsafe.Sizeof(uint(0)) > unsafe.Sizeof(uintptr(0)) {
			return uint64(*(*uint)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint32:
		if v.isIndirect() || unsafe.Sizeof(uint32(0)) > unsafe.Sizeof(uintptr(0)) {
			return uint64(*(*uint32)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	case Uint64:
		if v.isIndirect() || unsafe.Sizeof(uint64(0)) > unsafe.Sizeof(uintptr(0)) {
			return uint64(*(*uint64)(v.value))
		} else {
			return uint64(uintptr(v.value))
		}
	default:
		panic(&ValueError{Method: "Uint", Kind: v.Kind()})
	}
}

// CanFloat reports whether Float can be used without panicking.
func (v Value) CanFloat() bool {
	switch v.Kind() {
	case Float32, Float64:
		return true
	default:
		return false
	}
}

func (v Value) Float32() float32 {
	switch v.Kind() {
	case Float32:
		if v.isIndirect() || unsafe.Sizeof(float32(0)) > unsafe.Sizeof(uintptr(0)) {
			// The float is stored as an external value on systems with 16-bit
			// pointers.
			return *(*float32)(v.value)
		} else {
			// The float is directly stored in the interface value on systems
			// with 32-bit and 64-bit pointers.
			return *(*float32)(unsafe.Pointer(&v.value))
		}

	case Float64:
		return float32(v.Float())

	}

	panic(&ValueError{Method: "Float", Kind: v.Kind()})
}

func (v Value) Float() float64 {
	switch v.Kind() {
	case Float32:
		if v.isIndirect() || unsafe.Sizeof(float32(0)) > unsafe.Sizeof(uintptr(0)) {
			// The float is stored as an external value on systems with 16-bit
			// pointers.
			return float64(*(*float32)(v.value))
		} else {
			// The float is directly stored in the interface value on systems
			// with 32-bit and 64-bit pointers.
			return float64(*(*float32)(unsafe.Pointer(&v.value)))
		}
	case Float64:
		if v.isIndirect() || unsafe.Sizeof(float64(0)) > unsafe.Sizeof(uintptr(0)) {
			// For systems with 16-bit and 32-bit pointers.
			return *(*float64)(v.value)
		} else {
			// The float is directly stored in the interface value on systems
			// with 64-bit pointers.
			return *(*float64)(unsafe.Pointer(&v.value))
		}
	default:
		panic(&ValueError{Method: "Float", Kind: v.Kind()})
	}
}

// CanComplex reports whether Complex can be used without panicking.
func (v Value) CanComplex() bool {
	switch v.Kind() {
	case Complex64, Complex128:
		return true
	default:
		return false
	}
}

func (v Value) Complex() complex128 {
	switch v.Kind() {
	case Complex64:
		if v.isIndirect() || unsafe.Sizeof(complex64(0)) > unsafe.Sizeof(uintptr(0)) {
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
		panic(&ValueError{Method: "Complex", Kind: v.Kind()})
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
		return "<" + v.typecode.String() + " Value>"
	}
}

func (v Value) Bytes() []byte {
	switch v.Kind() {
	case Slice:
		if v.typecode.elem().Kind() != Uint8 {
			panic(&ValueError{Method: "Bytes", Kind: v.Kind()})
		}
		return *(*[]byte)(v.value)

	case Array:
		v.checkAddressable()

		if v.typecode.elem().Kind() != Uint8 {
			panic(&ValueError{Method: "Bytes", Kind: v.Kind()})
		}

		// Small inline arrays are not addressable, so we only have to
		// handle addressable arrays which will be stored as pointers
		// in v.value
		return unsafe.Slice((*byte)(v.value), v.Len())
	}

	panic(&ValueError{Method: "Bytes", Kind: v.Kind()})
}

func (v Value) Slice(i, j int) Value {
	switch v.Kind() {
	case Slice:
		hdr := *(*sliceHeader)(v.value)
		i, j := uintptr(i), uintptr(j)

		if j < i || hdr.cap < j {
			slicePanic()
		}

		elemSize := v.typecode.underlying().elem().Size()

		hdr.len = j - i
		hdr.cap = hdr.cap - i
		hdr.data = unsafe.Add(hdr.data, i*elemSize)

		return Value{
			typecode: v.typecode,
			value:    unsafe.Pointer(&hdr),
			flags:    v.flags,
		}

	case Array:
		// TODO(dgryski): can't do this yet because the resulting value needs type slice of v.elem(), not array of v.elem().
		// need to be able to look up this "new" type so pointer equality of types still works

	case String:
		i, j := uintptr(i), uintptr(j)
		str := *(*stringHeader)(v.value)

		if j < i || str.len < j {
			slicePanic()
		}

		hdr := stringHeader{
			data: unsafe.Add(str.data, i),
			len:  j - i,
		}

		return Value{
			typecode: v.typecode,
			value:    unsafe.Pointer(&hdr),
			flags:    v.flags,
		}
	}

	panic(&ValueError{Method: "Slice", Kind: v.Kind()})
}

func (v Value) Slice3(i, j, k int) Value {
	switch v.Kind() {
	case Slice:
		hdr := *(*sliceHeader)(v.value)
		i, j, k := uintptr(i), uintptr(j), uintptr(k)

		if j < i || k < j || hdr.len < k {
			slicePanic()
		}

		elemSize := v.typecode.underlying().elem().Size()

		hdr.len = j - i
		hdr.cap = k - i
		hdr.data = unsafe.Add(hdr.data, i*elemSize)

		return Value{
			typecode: v.typecode,
			value:    unsafe.Pointer(&hdr),
			flags:    v.flags,
		}

	case Array:
		// TODO(dgryski): can't do this yet because the resulting value needs type v.elem(), not array of v.elem().
		// need to be able to look up this "new" type so pointer equality of types still works

	}

	panic("unimplemented: (reflect.Value).Slice3()")
}

//go:linkname maplen runtime.hashmapLenUnsafePointer
func maplen(p unsafe.Pointer) int

//go:linkname chanlen runtime.chanLenUnsafePointer
func chanlen(p unsafe.Pointer) int

// Len returns the length of this value for slices, strings, arrays, channels,
// and maps. For other types, it panics.
func (v Value) Len() int {
	switch v.typecode.Kind() {
	case Array:
		return v.typecode.Len()
	case Chan:
		return chanlen(v.pointer())
	case Map:
		return maplen(v.pointer())
	case Slice:
		return int((*sliceHeader)(v.value).len)
	case String:
		return int((*stringHeader)(v.value).len)
	default:
		panic(&ValueError{Method: "Len", Kind: v.Kind()})
	}
}

//go:linkname chancap runtime.chanCapUnsafePointer
func chancap(p unsafe.Pointer) int

// Cap returns the capacity of this value for arrays, channels and slices.
// For other types, it panics.
func (v Value) Cap() int {
	switch v.typecode.Kind() {
	case Array:
		return v.typecode.Len()
	case Chan:
		return chancap(v.pointer())
	case Slice:
		return int((*sliceHeader)(v.value).cap)
	default:
		panic(&ValueError{Method: "Cap", Kind: v.Kind()})
	}
}

// NumField returns the number of fields of this struct. It panics for other
// value types.
func (v Value) NumField() int {
	return v.typecode.NumField()
}

func (v Value) Elem() Value {
	switch v.Kind() {
	case Ptr:
		ptr := v.pointer()
		if ptr == nil {
			return Value{}
		}
		// Don't copy RO flags
		flags := (v.flags & (valueFlagIndirect | valueFlagExported)) | valueFlagIndirect
		return Value{
			typecode: v.typecode.elem(),
			value:    ptr,
			flags:    flags,
		}
	case Interface:
		typecode, value := decomposeInterface(*(*interface{})(v.value))
		return Value{
			typecode: (*rawType)(typecode),
			value:    value,
			flags:    v.flags &^ valueFlagIndirect,
		}
	default:
		panic(&ValueError{Method: "Elem", Kind: v.Kind()})
	}
}

// Field returns the value of the i'th field of this struct.
func (v Value) Field(i int) Value {
	if v.Kind() != Struct {
		panic(&ValueError{Method: "Field", Kind: v.Kind()})
	}
	structField := v.typecode.rawField(i)

	// Copy flags but clear EmbedRO; we're not an embedded field anymore
	flags := v.flags & ^valueFlagEmbedRO
	if structField.PkgPath != "" {
		// No PkgPath => not exported.
		// Clear exported flag even if the parent was exported.
		flags &^= valueFlagExported

		// Update the RO flag
		if structField.Anonymous {
			// Embedded field
			flags |= valueFlagEmbedRO
		} else {
			flags |= valueFlagStickyRO
		}
	} else {
		// Parent field may not have been exported but we are
		flags |= valueFlagExported
	}

	size := v.typecode.Size()
	fieldType := structField.Type
	fieldSize := fieldType.Size()
	if v.isIndirect() || fieldSize > unsafe.Sizeof(uintptr(0)) {
		// v.value was already a pointer to the value and it should stay that
		// way.
		return Value{
			flags:    flags,
			typecode: fieldType,
			value:    unsafe.Add(v.value, structField.Offset),
		}
	}

	// The fieldSize is smaller than uintptr, which means that the value will
	// have to be stored directly in the interface value.

	if fieldSize == 0 {
		// The struct field is zero sized.
		// This is a rare situation, but because it's undefined behavior
		// to shift the size of the value (zeroing the value), handle this
		// situation explicitly.
		return Value{
			flags:    flags,
			typecode: fieldType,
			value:    unsafe.Pointer(nil),
		}
	}

	if size > unsafe.Sizeof(uintptr(0)) {
		// The value was not stored in the interface before but will be
		// afterwards, so load the value (from the correct offset) and return
		// it.
		ptr := unsafe.Add(v.value, structField.Offset)
		value := unsafe.Pointer(loadValue(ptr, fieldSize))
		return Value{
			flags:    flags &^ valueFlagIndirect,
			typecode: fieldType,
			value:    value,
		}
	}

	// The value was already stored directly in the interface and it still
	// is. Cut out the part of the value that we need.
	value := maskAndShift(uintptr(v.value), structField.Offset, fieldSize)
	return Value{
		flags:    flags,
		typecode: fieldType,
		value:    unsafe.Pointer(value),
	}
}

var uint8Type = TypeOf(uint8(0)).(*rawType)

func (v Value) Index(i int) Value {
	switch v.Kind() {
	case Slice:
		// Extract an element from the slice.
		slice := *(*sliceHeader)(v.value)
		if uint(i) >= uint(slice.len) {
			panic("reflect: slice index out of range")
		}
		flags := (v.flags & (valueFlagExported | valueFlagIndirect)) | valueFlagIndirect | v.flags.ro()
		elem := Value{
			typecode: v.typecode.elem(),
			flags:    flags,
		}
		elem.value = unsafe.Add(slice.data, elem.typecode.Size()*uintptr(i)) // pointer to new value
		return elem
	case String:
		// Extract a character from a string.
		// A string is never stored directly in the interface, but always as a
		// pointer to the string value.
		// Keeping valueFlagExported if set, but don't set valueFlagIndirect
		// otherwise CanSet will return true for string elements (which is bad,
		// strings are read-only).
		s := *(*stringHeader)(v.value)
		if uint(i) >= uint(s.len) {
			panic("reflect: string index out of range")
		}
		return Value{
			typecode: uint8Type,
			value:    unsafe.Pointer(uintptr(*(*uint8)(unsafe.Add(s.data, i)))),
			flags:    v.flags & valueFlagExported,
		}
	case Array:
		// Extract an element from the array.
		elemType := v.typecode.elem()
		elemSize := elemType.Size()
		size := v.typecode.Size()
		if size == 0 {
			// The element size is 0 and/or the length of the array is 0.
			return Value{
				typecode: v.typecode.elem(),
				flags:    v.flags,
			}
		}
		if elemSize > unsafe.Sizeof(uintptr(0)) {
			// The resulting value doesn't fit in a pointer so must be
			// indirect. Also, because size != 0 this implies that the array
			// length must be != 0, and thus that the total size is at least
			// elemSize.
			addr := unsafe.Add(v.value, elemSize*uintptr(i)) // pointer to new value
			return Value{
				typecode: v.typecode.elem(),
				flags:    v.flags,
				value:    addr,
			}
		}

		if size > unsafe.Sizeof(uintptr(0)) || v.isIndirect() {
			// The element fits in a pointer, but the array is not stored in the pointer directly.
			// Load the value from the pointer.
			addr := unsafe.Add(v.value, elemSize*uintptr(i)) // pointer to new value
			value := addr
			if !v.isIndirect() {
				// Use a pointer to the value (don't load the value) if the
				// 'indirect' flag is set.
				value = unsafe.Pointer(loadValue(addr, elemSize))
			}
			return Value{
				typecode: v.typecode.elem(),
				flags:    v.flags,
				value:    value,
			}
		}

		// The value fits in a pointer, so extract it with some shifting and
		// masking.
		offset := elemSize * uintptr(i)
		value := maskAndShift(uintptr(v.value), offset, elemSize)
		return Value{
			typecode: v.typecode.elem(),
			flags:    v.flags,
			value:    unsafe.Pointer(value),
		}
	default:
		panic(&ValueError{Method: "Index", Kind: v.Kind()})
	}
}

// loadValue loads a value that may or may not be word-aligned. The number of
// bytes given in size are loaded. The biggest possible size it can load is that
// of an uintptr.
func loadValue(ptr unsafe.Pointer, size uintptr) uintptr {
	loadedValue := uintptr(0)
	shift := uintptr(0)
	for i := uintptr(0); i < size; i++ {
		loadedValue |= uintptr(*(*byte)(ptr)) << shift
		shift += 8
		ptr = unsafe.Add(ptr, 1)
	}
	return loadedValue
}

// maskAndShift cuts out a part of a uintptr. Note that the offset may not be 0.
func maskAndShift(value, offset, size uintptr) uintptr {
	mask := ^uintptr(0) >> ((unsafe.Sizeof(uintptr(0)) - size) * 8)
	return (uintptr(value) >> (offset * 8)) & mask
}

func (v Value) NumMethod() int {
	return v.typecode.NumMethod()
}

// OverflowFloat reports whether the float64 x cannot be represented by v's type.
// It panics if v's Kind is not Float32 or Float64.
func (v Value) OverflowFloat(x float64) bool {
	k := v.Kind()
	switch k {
	case Float32:
		return overflowFloat32(x)
	case Float64:
		return false
	}
	panic(&ValueError{Method: "reflect.Value.OverflowFloat", Kind: v.Kind()})
}

func overflowFloat32(x float64) bool {
	if x < 0 {
		x = -x
	}
	return math.MaxFloat32 < x && x <= math.MaxFloat64
}

func (v Value) MapKeys() []Value {
	if v.Kind() != Map {
		panic(&ValueError{Method: "MapKeys", Kind: v.Kind()})
	}

	// empty map
	if v.Len() == 0 {
		return nil
	}

	keys := make([]Value, 0, v.Len())

	it := hashmapNewIterator()
	k := New(v.typecode.Key())
	e := New(v.typecode.Elem())

	keyType := v.typecode.key()
	isKeyStoredAsInterface := keyType.Kind() != String && !keyType.isBinary()

	for hashmapNext(v.pointer(), it, k.value, e.value) {
		if isKeyStoredAsInterface {
			intf := *(*interface{})(k.value)
			v := ValueOf(intf)
			keys = append(keys, v)
		} else {
			keys = append(keys, k.Elem())
		}
		k = New(v.typecode.Key())
	}

	return keys
}

//go:linkname hashmapStringGet runtime.hashmapStringGetUnsafePointer
func hashmapStringGet(m unsafe.Pointer, key string, value unsafe.Pointer, valueSize uintptr) bool

//go:linkname hashmapBinaryGet runtime.hashmapBinaryGetUnsafePointer
func hashmapBinaryGet(m unsafe.Pointer, key, value unsafe.Pointer, valueSize uintptr) bool

//go:linkname hashmapInterfaceGet runtime.hashmapInterfaceGetUnsafePointer
func hashmapInterfaceGet(m unsafe.Pointer, key interface{}, value unsafe.Pointer, valueSize uintptr) bool

func (v Value) MapIndex(key Value) Value {
	if v.Kind() != Map {
		panic(&ValueError{Method: "MapIndex", Kind: v.Kind()})
	}

	vkey := v.typecode.key()

	// compare key type with actual key type of map
	if !key.typecode.AssignableTo(vkey) {
		// type error?
		panic("reflect.Value.MapIndex: incompatible types for key")
	}

	elemType := v.typecode.Elem()
	elem := New(elemType)

	if vkey.Kind() == String {
		if ok := hashmapStringGet(v.pointer(), *(*string)(key.value), elem.value, elemType.Size()); !ok {
			return Value{}
		}
		return elem.Elem()
	} else if vkey.isBinary() {
		var keyptr unsafe.Pointer
		if key.isIndirect() || key.typecode.Size() > unsafe.Sizeof(uintptr(0)) {
			keyptr = key.value
		} else {
			keyptr = unsafe.Pointer(&key.value)
		}
		//TODO(dgryski): zero out padding bytes in key, if any
		if ok := hashmapBinaryGet(v.pointer(), keyptr, elem.value, elemType.Size()); !ok {
			return Value{}
		}
		return elem.Elem()
	} else {
		if ok := hashmapInterfaceGet(v.pointer(), key.Interface(), elem.value, elemType.Size()); !ok {
			return Value{}
		}
		return elem.Elem()
	}
}

//go:linkname hashmapNewIterator runtime.hashmapNewIterator
func hashmapNewIterator() unsafe.Pointer

//go:linkname hashmapNext runtime.hashmapNextUnsafePointer
func hashmapNext(m unsafe.Pointer, it unsafe.Pointer, key, value unsafe.Pointer) bool

func (v Value) MapRange() *MapIter {
	if v.Kind() != Map {
		panic(&ValueError{Method: "MapRange", Kind: v.Kind()})
	}

	keyType := v.typecode.key()
	isKeyStoredAsInterface := keyType.Kind() != String && !keyType.isBinary()

	return &MapIter{
		m:            v,
		it:           hashmapNewIterator(),
		keyInterface: isKeyStoredAsInterface,
	}
}

type MapIter struct {
	m   Value
	it  unsafe.Pointer
	key Value
	val Value

	valid        bool
	keyInterface bool
}

func (it *MapIter) Key() Value {
	if !it.valid {
		panic("reflect.MapIter.Key called on invalid iterator")
	}

	if it.keyInterface {
		intf := *(*interface{})(it.key.value)
		v := ValueOf(intf)
		return v
	}

	return it.key.Elem()
}

func (it *MapIter) Value() Value {
	if !it.valid {
		panic("reflect.MapIter.Value called on invalid iterator")
	}

	return it.val.Elem()
}

func (it *MapIter) Next() bool {
	it.key = New(it.m.typecode.Key())
	it.val = New(it.m.typecode.Elem())

	it.valid = hashmapNext(it.m.pointer(), it.it, it.key.value, it.val.value)
	return it.valid
}

func (v Value) Set(x Value) {
	v.checkAddressable()
	v.checkRO()
	if !x.typecode.AssignableTo(v.typecode) {
		panic("reflect: cannot set")
	}

	if v.typecode.Kind() == Interface && x.typecode.Kind() != Interface {
		intf := composeInterface(unsafe.Pointer(x.typecode), x.value)
		x = Value{
			typecode: v.typecode,
			value:    unsafe.Pointer(&intf),
		}
	}

	size := v.typecode.Size()
	xptr := x.value
	if size <= unsafe.Sizeof(uintptr(0)) && !x.isIndirect() {
		value := x.value
		xptr = unsafe.Pointer(&value)
	}
	memcpy(v.value, xptr, size)
}

func (v Value) SetBool(x bool) {
	v.checkAddressable()
	v.checkRO()
	switch v.Kind() {
	case Bool:
		*(*bool)(v.value) = x
	default:
		panic(&ValueError{Method: "SetBool", Kind: v.Kind()})
	}
}

func (v Value) SetInt(x int64) {
	v.checkAddressable()
	v.checkRO()
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
		panic(&ValueError{Method: "SetInt", Kind: v.Kind()})
	}
}

func (v Value) SetUint(x uint64) {
	v.checkAddressable()
	v.checkRO()
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
		panic(&ValueError{Method: "SetUint", Kind: v.Kind()})
	}
}

func (v Value) SetFloat(x float64) {
	v.checkAddressable()
	v.checkRO()
	switch v.Kind() {
	case Float32:
		*(*float32)(v.value) = float32(x)
	case Float64:
		*(*float64)(v.value) = x
	default:
		panic(&ValueError{Method: "SetFloat", Kind: v.Kind()})
	}
}

func (v Value) SetComplex(x complex128) {
	v.checkAddressable()
	v.checkRO()
	switch v.Kind() {
	case Complex64:
		*(*complex64)(v.value) = complex64(x)
	case Complex128:
		*(*complex128)(v.value) = x
	default:
		panic(&ValueError{Method: "SetComplex", Kind: v.Kind()})
	}
}

func (v Value) SetString(x string) {
	v.checkAddressable()
	v.checkRO()
	switch v.Kind() {
	case String:
		*(*string)(v.value) = x
	default:
		panic(&ValueError{Method: "SetString", Kind: v.Kind()})
	}
}

func (v Value) SetBytes(x []byte) {
	v.checkAddressable()
	v.checkRO()
	if v.typecode.Kind() != Slice || v.typecode.elem().Kind() != Uint8 {
		panic("reflect.Value.SetBytes called on not []byte")
	}

	// copy the header contents over
	*(*[]byte)(v.value) = x
}

func (v Value) SetCap(n int) {
	panic("unimplemented: (reflect.Value).SetCap()")
}

func (v Value) SetLen(n int) {
	if v.typecode.Kind() != Slice {
		panic(&ValueError{Method: "reflect.Value.SetLen", Kind: v.Kind()})
	}
	v.checkAddressable()
	hdr := (*sliceHeader)(v.value)
	if int(uintptr(n)) != n || uintptr(n) > hdr.cap {
		panic("reflect.Value.SetLen: slice length out of range")
	}
	hdr.len = uintptr(n)
}

func (v Value) checkAddressable() {
	if !v.isIndirect() {
		panic("reflect: value is not addressable")
	}
}

// OverflowInt reports whether the int64 x cannot be represented by v's type.
// It panics if v's Kind is not Int, Int8, Int16, Int32, or Int64.
func (v Value) OverflowInt(x int64) bool {
	switch v.Kind() {
	case Int, Int8, Int16, Int32, Int64:
		bitSize := v.typecode.Size() * 8
		trunc := (x << (64 - bitSize)) >> (64 - bitSize)
		return x != trunc
	}
	panic(&ValueError{Method: "reflect.Value.OverflowInt", Kind: v.Kind()})
}

// OverflowUint reports whether the uint64 x cannot be represented by v's type.
// It panics if v's Kind is not Uint, Uintptr, Uint8, Uint16, Uint32, or Uint64.
func (v Value) OverflowUint(x uint64) bool {
	k := v.Kind()
	switch k {
	case Uint, Uintptr, Uint8, Uint16, Uint32, Uint64:
		bitSize := v.typecode.Size() * 8
		trunc := (x << (64 - bitSize)) >> (64 - bitSize)
		return x != trunc
	}
	panic(&ValueError{Method: "reflect.Value.OverflowUint", Kind: v.Kind()})
}

func (v Value) CanConvert(t Type) bool {
	panic("unimplemented: (reflect.Value).CanConvert()")
}

func (v Value) Convert(t Type) Value {
	if v, ok := convertOp(v, t); ok {
		return v
	}

	panic("reflect.Value.Convert: value of type " + v.typecode.String() + " cannot be converted to type " + t.String())
}

func convertOp(src Value, typ Type) (Value, bool) {

	// Easy check first.  Do we even need to do anything?
	if src.typecode.underlying() == typ.(*rawType).underlying() {
		return Value{
			typecode: typ.(*rawType),
			value:    src.value,
			flags:    src.flags,
		}, true
	}

	switch src.Kind() {
	case Int, Int8, Int16, Int32, Int64:
		switch rtype := typ.(*rawType); rtype.Kind() {
		case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
			return cvtInt(src, rtype), true
		case Float32, Float64:
			return cvtIntFloat(src, rtype), true
		case String:
			return cvtIntString(src, rtype), true
		}

	case Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
		switch rtype := typ.(*rawType); rtype.Kind() {
		case Int, Int8, Int16, Int32, Int64, Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
			return cvtUint(src, rtype), true
		case Float32, Float64:
			return cvtUintFloat(src, rtype), true
		case String:
			return cvtUintString(src, rtype), true
		}

	case Float32, Float64:
		switch rtype := typ.(*rawType); rtype.Kind() {
		case Int, Int8, Int16, Int32, Int64:
			return cvtFloatInt(src, rtype), true
		case Uint, Uint8, Uint16, Uint32, Uint64, Uintptr:
			return cvtFloatUint(src, rtype), true
		case Float32, Float64:
			return cvtFloat(src, rtype), true
		}

		/*
			case Complex64, Complex128:
				switch src.Kind() {
				case Complex64, Complex128:
					return cvtComplex
				}
		*/

	case Slice:
		if typ.Kind() == String && !src.typecode.elem().isNamed() {
			rtype := typ.(*rawType)

			switch src.Type().Elem().Kind() {
			case Uint8:
				return cvtBytesString(src, rtype), true
			case Int32:
				return cvtRunesString(src, rtype), true
			}
		}

	case String:
		rtype := typ.(*rawType)
		if typ.Kind() == Slice && !rtype.elem().isNamed() {
			switch typ.Elem().Kind() {
			case Uint8:
				return cvtStringBytes(src, rtype), true
			case Int32:
				return cvtStringRunes(src, rtype), true
			}
		}
	}

	// TODO(dgryski): Unimplemented:
	// Chan
	// Non-defined pointers types with same underlying base type
	// Interface <-> Type conversions

	return Value{}, false
}

func cvtInt(v Value, t *rawType) Value {
	return makeInt(v.flags, uint64(v.Int()), t)
}

func cvtUint(v Value, t *rawType) Value {
	return makeInt(v.flags, v.Uint(), t)
}

func cvtIntFloat(v Value, t *rawType) Value {
	return makeFloat(v.flags, float64(v.Int()), t)
}

func cvtUintFloat(v Value, t *rawType) Value {
	return makeFloat(v.flags, float64(v.Uint()), t)
}

func cvtFloatInt(v Value, t *rawType) Value {
	return makeInt(v.flags, uint64(int64(v.Float())), t)
}

func cvtFloatUint(v Value, t *rawType) Value {
	return makeInt(v.flags, uint64(v.Float()), t)
}

func cvtFloat(v Value, t *rawType) Value {
	if v.Type().Kind() == Float32 && t.Kind() == Float32 {
		// Don't do any conversion if both types have underlying type float32.
		// This avoids converting to float64 and back, which will
		// convert a signaling NaN to a quiet NaN. See issue 36400.
		return makeFloat32(v.flags, v.Float32(), t)
	}
	return makeFloat(v.flags, v.Float(), t)
}

//go:linkname stringToBytes runtime.stringToBytes
func stringToBytes(x string) []byte

func cvtStringBytes(v Value, t *rawType) Value {
	b := stringToBytes(*(*string)(v.value))
	return Value{
		typecode: t,
		value:    unsafe.Pointer(&b),
		flags:    v.flags,
	}
}

//go:linkname stringFromBytes runtime.stringFromBytes
func stringFromBytes(x []byte) string

func cvtBytesString(v Value, t *rawType) Value {
	s := stringFromBytes(*(*[]byte)(v.value))
	return Value{
		typecode: t,
		value:    unsafe.Pointer(&s),
		flags:    v.flags,
	}
}

func makeInt(flags valueFlags, bits uint64, t *rawType) Value {
	size := t.Size()

	v := Value{
		typecode: t,
		flags:    flags,
	}

	ptr := unsafe.Pointer(&v.value)
	if size > unsafe.Sizeof(uintptr(0)) {
		ptr = alloc(size, nil)
		v.value = ptr
	}

	switch size {
	case 1:
		*(*uint8)(ptr) = uint8(bits)
	case 2:
		*(*uint16)(ptr) = uint16(bits)
	case 4:
		*(*uint32)(ptr) = uint32(bits)
	case 8:
		*(*uint64)(ptr) = bits
	}
	return v
}

func makeFloat(flags valueFlags, f float64, t *rawType) Value {
	size := t.Size()

	v := Value{
		typecode: t,
		flags:    flags,
	}

	ptr := unsafe.Pointer(&v.value)
	if size > unsafe.Sizeof(uintptr(0)) {
		ptr = alloc(size, nil)
		v.value = ptr
	}

	switch size {
	case 4:
		*(*float32)(ptr) = float32(f)
	case 8:
		*(*float64)(ptr) = f
	}
	return v
}

func makeFloat32(flags valueFlags, f float32, t *rawType) Value {
	v := Value{
		typecode: t,
		flags:    flags,
	}
	*(*float32)(unsafe.Pointer(&v.value)) = float32(f)
	return v
}

func cvtIntString(src Value, t *rawType) Value {
	panic("cvtUintString: unimplemented")
}

func cvtUintString(src Value, t *rawType) Value {
	panic("cvtUintString: unimplemented")
}

func cvtStringRunes(src Value, t *rawType) Value {
	panic("cvsStringRunes: unimplemented")
}

func cvtRunesString(src Value, t *rawType) Value {
	panic("cvsRunesString: unimplemented")
}

//go:linkname slicePanic runtime.slicePanic
func slicePanic()

func MakeSlice(typ Type, len, cap int) Value {
	if typ.Kind() != Slice {
		panic("reflect.MakeSlice of non-slice type")
	}

	rtype := typ.(*rawType)

	ulen := uint(len)
	ucap := uint(cap)
	maxSize := (^uintptr(0)) / 2
	elementSize := rtype.elem().Size()
	if elementSize > 1 {
		maxSize /= uintptr(elementSize)
	}
	if ulen > ucap || ucap > uint(maxSize) {
		slicePanic()
	}

	// This can't overflow because of the above checks.
	size := uintptr(ucap) * elementSize

	var slice sliceHeader
	slice.cap = uintptr(ucap)
	slice.len = uintptr(ulen)
	slice.data = alloc(size, nil)

	return Value{
		typecode: rtype,
		value:    unsafe.Pointer(&slice),
		flags:    valueFlagExported,
	}
}

var zerobuffer unsafe.Pointer

const zerobufferLen = 32

func init() {
	// 32 characters of zero bytes
	zerobufferStr := "\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00\x00"
	s := (*stringHeader)(unsafe.Pointer(&zerobufferStr))
	zerobuffer = s.data
}

func Zero(typ Type) Value {
	size := typ.Size()
	if size <= unsafe.Sizeof(uintptr(0)) {
		return Value{
			typecode: typ.(*rawType),
			value:    nil,
			flags:    valueFlagExported | valueFlagRO,
		}
	}

	if size <= zerobufferLen {
		return Value{
			typecode: typ.(*rawType),
			value:    unsafe.Pointer(zerobuffer),
			flags:    valueFlagExported | valueFlagRO,
		}
	}

	return Value{
		typecode: typ.(*rawType),
		value:    alloc(size, nil),
		flags:    valueFlagExported | valueFlagRO,
	}
}

// New is the reflect equivalent of the new(T) keyword, returning a pointer to a
// new value of the given type.
func New(typ Type) Value {
	return Value{
		typecode: pointerTo(typ.(*rawType)),
		value:    alloc(typ.Size(), nil),
		flags:    valueFlagExported,
	}
}

type funcHeader struct {
	Context unsafe.Pointer
	Code    unsafe.Pointer
}

type SliceHeader struct {
	Data uintptr
	Len  uintptr
	Cap  uintptr
}

// Slice header that matches the underlying structure. Used for when we switch
// to a precise GC, which needs to know exactly where pointers live.
type sliceHeader struct {
	data unsafe.Pointer
	len  uintptr
	cap  uintptr
}

type StringHeader struct {
	Data uintptr
	Len  uintptr
}

// Like sliceHeader, this type is used internally to make sure pointer and
// non-pointer fields match those of actual strings.
type stringHeader struct {
	data unsafe.Pointer
	len  uintptr
}

type ValueError struct {
	Method string
	Kind   Kind
}

func (e *ValueError) Error() string {
	if e.Kind == 0 {
		return "reflect: call of " + e.Method + " on zero Value"
	}
	return "reflect: call of " + e.Method + " on " + e.Kind.String() + " Value"
}

//go:linkname memcpy runtime.memcpy
func memcpy(dst, src unsafe.Pointer, size uintptr)

//go:linkname alloc runtime.alloc
func alloc(size uintptr, layout unsafe.Pointer) unsafe.Pointer

//go:linkname sliceAppend runtime.sliceAppend
func sliceAppend(srcBuf, elemsBuf unsafe.Pointer, srcLen, srcCap, elemsLen uintptr, elemSize uintptr) (unsafe.Pointer, uintptr, uintptr)

//go:linkname sliceCopy runtime.sliceCopy
func sliceCopy(dst, src unsafe.Pointer, dstLen, srcLen uintptr, elemSize uintptr) int

// Copy copies the contents of src into dst until either
// dst has been filled or src has been exhausted.
func Copy(dst, src Value) int {
	compatibleTypes := false ||
		// dst and src are both slices or arrays with equal types
		((dst.typecode.Kind() == Slice || dst.typecode.Kind() == Array) &&
			(src.typecode.Kind() == Slice || src.typecode.Kind() == Array) &&
			(dst.typecode.elem() == src.typecode.elem())) ||
		// dst is array or slice of uint8 and src is string
		((dst.typecode.Kind() == Slice || dst.typecode.Kind() == Array) &&
			dst.typecode.elem().Kind() == Uint8 &&
			src.typecode.Kind() == String)

	if !compatibleTypes {
		panic("Copy: type mismatch: " + dst.typecode.String() + "/" + src.typecode.String())
	}

	// Can read from an unaddressable array but not write to one.
	if dst.typecode.Kind() == Array && !dst.isIndirect() {
		panic("reflect.Copy: unaddressable array value")
	}

	dstbuf, dstlen := buflen(dst)
	srcbuf, srclen := buflen(src)

	if srclen > 0 {
		dst.checkRO()
	}

	return sliceCopy(dstbuf, srcbuf, dstlen, srclen, dst.typecode.elem().Size())
}

func buflen(v Value) (unsafe.Pointer, uintptr) {
	var buf unsafe.Pointer
	var len uintptr
	switch v.typecode.Kind() {
	case Slice:
		hdr := (*sliceHeader)(v.value)
		buf = hdr.data
		len = hdr.len
	case Array:
		if v.isIndirect() {
			buf = v.value
		} else {
			buf = unsafe.Pointer(&v.value)
		}
		len = uintptr(v.Len())
	case String:
		hdr := (*stringHeader)(v.value)
		buf = hdr.data
		len = hdr.len
	default:
		// This shouldn't happen
		panic("reflect.Copy: not slice or array or string")
	}

	return buf, len
}

//go:linkname sliceGrow runtime.sliceGrow
func sliceGrow(buf unsafe.Pointer, oldLen, oldCap, newCap, elemSize uintptr) (unsafe.Pointer, uintptr, uintptr)

// extend slice to hold n new elements
func (v *Value) extendSlice(n int) {
	if v.Kind() != Slice {
		panic(&ValueError{Method: "extendSlice", Kind: v.Kind()})
	}

	var old sliceHeader
	if v.value != nil {
		old = *(*sliceHeader)(v.value)
	}

	var nbuf unsafe.Pointer
	var nlen, ncap uintptr

	if old.len+uintptr(n) > old.cap {
		// we need to grow the slice
		nbuf, nlen, ncap = sliceGrow(old.data, old.len, old.cap, old.cap+uintptr(n), v.typecode.elem().Size())
	} else {
		// we can reuse the slice we have
		nbuf = old.data
		nlen = old.len
		ncap = old.cap
	}

	newslice := sliceHeader{
		data: nbuf,
		len:  nlen + uintptr(n),
		cap:  ncap,
	}

	v.flags = valueFlagExported
	v.value = (unsafe.Pointer)(&newslice)
}

// Append appends the values x to a slice s and returns the resulting slice.
// As in Go, each x's value must be assignable to the slice's element type.
func Append(v Value, x ...Value) Value {
	if v.Kind() != Slice {
		panic(&ValueError{Method: "Append", Kind: v.Kind()})
	}
	oldLen := v.Len()
	v.extendSlice(len(x))
	for i, xx := range x {
		v.Index(oldLen + i).Set(xx)
	}
	return v
}

// AppendSlice appends a slice t to a slice s and returns the resulting slice.
// The slices s and t must have the same element type.
func AppendSlice(s, t Value) Value {
	if s.typecode.Kind() != Slice || t.typecode.Kind() != Slice || s.typecode != t.typecode {
		// Not a very helpful error message, but shortened to just one error to
		// keep code size down.
		panic("reflect.AppendSlice: invalid types")
	}
	if !s.isExported() || !t.isExported() {
		// One of the sides was not exported, so can't access the data.
		panic("reflect.AppendSlice: unexported")
	}
	sSlice := (*sliceHeader)(s.value)
	tSlice := (*sliceHeader)(t.value)
	elemSize := s.typecode.elem().Size()
	ptr, len, cap := sliceAppend(sSlice.data, tSlice.data, sSlice.len, sSlice.cap, tSlice.len, elemSize)
	result := &sliceHeader{
		data: ptr,
		len:  len,
		cap:  cap,
	}
	return Value{
		typecode: s.typecode,
		value:    unsafe.Pointer(result),
		flags:    valueFlagExported,
	}
}

//go:linkname hashmapStringSet runtime.hashmapStringSetUnsafePointer
func hashmapStringSet(m unsafe.Pointer, key string, value unsafe.Pointer)

//go:linkname hashmapBinarySet runtime.hashmapBinarySetUnsafePointer
func hashmapBinarySet(m unsafe.Pointer, key, value unsafe.Pointer)

//go:linkname hashmapInterfaceSet runtime.hashmapInterfaceSetUnsafePointer
func hashmapInterfaceSet(m unsafe.Pointer, key interface{}, value unsafe.Pointer)

//go:linkname hashmapStringDelete runtime.hashmapStringDeleteUnsafePointer
func hashmapStringDelete(m unsafe.Pointer, key string)

//go:linkname hashmapBinaryDelete runtime.hashmapBinaryDeleteUnsafePointer
func hashmapBinaryDelete(m unsafe.Pointer, key unsafe.Pointer)

//go:linkname hashmapInterfaceDelete runtime.hashmapInterfaceDeleteUnsafePointer
func hashmapInterfaceDelete(m unsafe.Pointer, key interface{})

func (v Value) SetMapIndex(key, elem Value) {
	v.checkRO()
	if v.Kind() != Map {
		panic(&ValueError{Method: "SetMapIndex", Kind: v.Kind()})
	}

	vkey := v.typecode.key()

	// compare key type with actual key type of map
	if !key.typecode.AssignableTo(vkey) {
		panic("reflect.Value.SetMapIndex: incompatible types for key")
	}

	// if elem is the zero Value, it means delete
	del := elem == Value{}

	if !del && !elem.typecode.AssignableTo(v.typecode.elem()) {
		panic("reflect.Value.SetMapIndex: incompatible types for value")
	}

	// make elem an interface if it needs to be converted
	if v.typecode.elem().Kind() == Interface && elem.typecode.Kind() != Interface {
		intf := composeInterface(unsafe.Pointer(elem.typecode), elem.value)
		elem = Value{
			typecode: v.typecode.elem(),
			value:    unsafe.Pointer(&intf),
		}
	}

	if key.Kind() == String {
		if del {
			hashmapStringDelete(v.pointer(), *(*string)(key.value))
		} else {
			var elemptr unsafe.Pointer
			if elem.isIndirect() || elem.typecode.Size() > unsafe.Sizeof(uintptr(0)) {
				elemptr = elem.value
			} else {
				elemptr = unsafe.Pointer(&elem.value)
			}
			hashmapStringSet(v.pointer(), *(*string)(key.value), elemptr)
		}

	} else if key.typecode.isBinary() {
		var keyptr unsafe.Pointer
		if key.isIndirect() || key.typecode.Size() > unsafe.Sizeof(uintptr(0)) {
			keyptr = key.value
		} else {
			keyptr = unsafe.Pointer(&key.value)
		}

		if del {
			hashmapBinaryDelete(v.pointer(), keyptr)
		} else {
			var elemptr unsafe.Pointer
			if elem.isIndirect() || elem.typecode.Size() > unsafe.Sizeof(uintptr(0)) {
				elemptr = elem.value
			} else {
				elemptr = unsafe.Pointer(&elem.value)
			}
			hashmapBinarySet(v.pointer(), keyptr, elemptr)
		}
	} else {
		if del {
			hashmapInterfaceDelete(v.pointer(), key.Interface())
		} else {
			var elemptr unsafe.Pointer
			if elem.isIndirect() || elem.typecode.Size() > unsafe.Sizeof(uintptr(0)) {
				elemptr = elem.value
			} else {
				elemptr = unsafe.Pointer(&elem.value)
			}

			hashmapInterfaceSet(v.pointer(), key.Interface(), elemptr)
		}
	}
}

// FieldByIndex returns the nested field corresponding to index.
func (v Value) FieldByIndex(index []int) Value {
	if len(index) == 1 {
		return v.Field(index[0])
	}
	if v.Kind() != Struct {
		panic(&ValueError{"FieldByIndex", v.Kind()})
	}
	for i, x := range index {
		if i > 0 {
			if v.Kind() == Pointer && v.typecode.elem().Kind() == Struct {
				if v.IsNil() {
					panic("reflect: indirection through nil pointer to embedded struct")
				}
				v = v.Elem()
			}
		}
		v = v.Field(x)
	}
	return v
}

// FieldByIndexErr returns the nested field corresponding to index.
func (v Value) FieldByIndexErr(index []int) (Value, error) {
	return Value{}, &ValueError{Method: "FieldByIndexErr"}
}

func (v Value) FieldByName(name string) Value {
	if v.Kind() != Struct {
		panic(&ValueError{"FieldByName", v.Kind()})
	}

	if field, ok := v.typecode.FieldByName(name); ok {
		return v.FieldByIndex(field.Index)
	}
	return Value{}
}

func (v Value) FieldByNameFunc(match func(string) bool) Value {
	if v.Kind() != Struct {
		panic(&ValueError{"FieldByName", v.Kind()})
	}

	if field, ok := v.typecode.FieldByNameFunc(match); ok {
		return v.FieldByIndex(field.Index)
	}
	return Value{}
}

//go:linkname hashmapMake runtime.hashmapMakeUnsafePointer
func hashmapMake(keySize, valueSize uintptr, sizeHint uintptr, alg uint8) unsafe.Pointer

// MakeMapWithSize creates a new map with the specified type and initial space
// for approximately n elements.
func MakeMapWithSize(typ Type, n int) Value {

	// TODO(dgryski): deduplicate these?  runtime and reflect both need them.
	const (
		hashmapAlgorithmBinary uint8 = iota
		hashmapAlgorithmString
		hashmapAlgorithmInterface
	)

	if typ.Kind() != Map {
		panic(&ValueError{Method: "MakeMap", Kind: typ.Kind()})
	}

	if n < 0 {
		panic("reflect.MakeMapWithSize: negative size hint")
	}

	key := typ.Key().(*rawType)
	val := typ.Elem().(*rawType)

	var alg uint8

	if key.Kind() == String {
		alg = hashmapAlgorithmString
	} else if key.isBinary() {
		alg = hashmapAlgorithmBinary
	} else {
		alg = hashmapAlgorithmInterface
	}

	m := hashmapMake(key.Size(), val.Size(), uintptr(n), alg)

	return Value{
		typecode: typ.(*rawType),
		value:    m,
		flags:    valueFlagExported,
	}
}

type SelectDir int

const (
	_             SelectDir = iota
	SelectSend              // case Chan <- Send
	SelectRecv              // case <-Chan:
	SelectDefault           // default
)

type SelectCase struct {
	Dir  SelectDir // direction of case
	Chan Value     // channel to use (for send or receive)
	Send Value     // value to send (for send)
}

func Select(cases []SelectCase) (chosen int, recv Value, recvOK bool) {
	panic("unimplemented: reflect.Select")
}

func (v Value) Send(x Value) {
	panic("unimplemented: reflect.Value.Send()")
}

func (v Value) Close() {
	panic("unimplemented: reflect.Value.Close()")
}

// MakeMap creates a new map with the specified type.
func MakeMap(typ Type) Value {
	return MakeMapWithSize(typ, 8)
}

func (v Value) Call(in []Value) []Value {
	panic("unimplemented: (reflect.Value).Call()")
}

func (v Value) Method(i int) Value {
	panic("unimplemented: (reflect.Value).Method()")
}

func (v Value) MethodByName(name string) Value {
	panic("unimplemented: (reflect.Value).MethodByName()")
}

func (v Value) Recv() (x Value, ok bool) {
	panic("unimplemented: (reflect.Value).Recv()")
}

func NewAt(typ Type, p unsafe.Pointer) Value {
	panic("unimplemented: reflect.New()")
}

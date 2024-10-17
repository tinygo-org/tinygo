package cm

import "unsafe"

// AnyInteger is a type constraint for any integer type.
type AnyInteger interface {
	~int | ~uint | ~uintptr | ~int8 | ~uint8 | ~int16 | ~uint16 | ~int32 | ~uint32 | ~int64 | ~uint64
}

// Reinterpret reinterprets the bits of type From into type T.
// Will panic if the size of From is smaller than the size of To.
func Reinterpret[T, From any](from From) (to T) {
	if unsafe.Sizeof(to) > unsafe.Sizeof(from) {
		panic("reinterpret: size of to > from")
	}
	return *(*T)(unsafe.Pointer(&from))
}

// LowerString lowers a [string] into a pair of Core WebAssembly types.
//
// [string]: https://pkg.go.dev/builtin#string
func LowerString[S ~string](s S) (*byte, uint32) {
	return unsafe.StringData(string(s)), uint32(len(s))
}

// LiftString lifts Core WebAssembly types into a [string].
func LiftString[T ~string, Data unsafe.Pointer | uintptr | *uint8, Len AnyInteger](data Data, len Len) T {
	return T(unsafe.String((*uint8)(unsafe.Pointer(data)), int(len)))
}

// LowerList lowers a [List] into a pair of Core WebAssembly types.
func LowerList[L AnyList[T], T any](list L) (*T, uint32) {
	l := (*List[T])(unsafe.Pointer(&list))
	return l.data, uint32(l.len)
}

// LiftList lifts Core WebAssembly types into a [List].
func LiftList[L AnyList[T], T any, Data unsafe.Pointer | uintptr | *T, Len AnyInteger](data Data, len Len) L {
	return L(NewList((*T)(unsafe.Pointer(data)), len))
}

// BoolToU32 converts a value whose underlying type is [bool] into a [uint32].
// Used to lower a [bool] into a Core WebAssembly i32 as specified in the [Canonical ABI].
//
// [bool]: https://pkg.go.dev/builtin#bool
// [uint32]: https://pkg.go.dev/builtin#uint32
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
func BoolToU32[B ~bool](v B) uint32 { return uint32(*(*uint8)(unsafe.Pointer(&v))) }

// U32ToBool converts a [uint32] into a [bool].
// Used to lift a Core WebAssembly i32 into a [bool] as specified in the [Canonical ABI].
//
// [uint32]: https://pkg.go.dev/builtin#uint32
// [bool]: https://pkg.go.dev/builtin#bool
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
func U32ToBool(v uint32) bool { tmp := uint8(v); return *(*bool)(unsafe.Pointer(&tmp)) }

// F32ToU32 maps the bits of a [float32] into a [uint32].
// Used to lower a [float32] into a Core WebAssembly i32 as specified in the [Canonical ABI].
//
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
// [float32]: https://pkg.go.dev/builtin#float32
// [uint32]: https://pkg.go.dev/builtin#uint32
func F32ToU32(v float32) uint32 { return *(*uint32)(unsafe.Pointer(&v)) }

// U32ToF32 maps the bits of a [uint32] into a [float32].
// Used to lift a Core WebAssembly i32 into a [float32] as specified in the [Canonical ABI].
//
// [uint32]: https://pkg.go.dev/builtin#uint32
// [float32]: https://pkg.go.dev/builtin#float32
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
func U32ToF32(v uint32) float32 { return *(*float32)(unsafe.Pointer(&v)) }

// F64ToU64 maps the bits of a [float64] into a [uint64].
// Used to lower a [float64] into a Core WebAssembly i64 as specified in the [Canonical ABI].
//
// [float64]: https://pkg.go.dev/builtin#float64
// [uint64]: https://pkg.go.dev/builtin#uint64
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
//
// [uint32]: https://pkg.go.dev/builtin#uint32
func F64ToU64(v float64) uint64 { return *(*uint64)(unsafe.Pointer(&v)) }

// U64ToF64 maps the bits of a [uint64] into a [float64].
// Used to lift a Core WebAssembly i64 into a [float64] as specified in the [Canonical ABI].
//
// [uint64]: https://pkg.go.dev/builtin#uint64
// [float64]: https://pkg.go.dev/builtin#float64
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
func U64ToF64(v uint64) float64 { return *(*float64)(unsafe.Pointer(&v)) }

// F32ToU64 maps the bits of a [float32] into a [uint64].
// Used to lower a [float32] into a Core WebAssembly i64 when required by the [Canonical ABI].
//
// [float32]: https://pkg.go.dev/builtin#float32
// [uint64]: https://pkg.go.dev/builtin#uint64
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
func F32ToU64(v float32) uint64 { return uint64(*(*uint32)(unsafe.Pointer(&v))) }

// U64ToF32 maps the bits of a [uint64] into a [float32].
// Used to lift a Core WebAssembly i64 into a [float32] when required by the [Canonical ABI].
//
// [uint64]: https://pkg.go.dev/builtin#uint64
// [float32]: https://pkg.go.dev/builtin#float32
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
func U64ToF32(v uint64) float32 {
	truncated := uint32(v)
	return *(*float32)(unsafe.Pointer(&truncated))
}

// PointerToU32 converts a pointer of type *T into a [uint32].
// Used to lower a pointer into a Core WebAssembly i32 as specified in the [Canonical ABI].
//
// [uint32]: https://pkg.go.dev/builtin#uint32
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
func PointerToU32[T any](v *T) uint32 { return uint32(uintptr(unsafe.Pointer(v))) }

// U32ToPointer converts a [uint32] into a pointer of type *T.
// Used to lift a Core WebAssembly i32 into a pointer as specified in the [Canonical ABI].
//
// [uint32]: https://pkg.go.dev/builtin#uint32
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
func U32ToPointer[T any](v uint32) *T { return (*T)(unsafePointer(uintptr(v))) }

// PointerToU64 converts a pointer of type *T into a [uint64].
// Used to lower a pointer into a Core WebAssembly i64 as specified in the [Canonical ABI].
//
// [uint64]: https://pkg.go.dev/builtin#uint64
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
func PointerToU64[T any](v *T) uint64 { return uint64(uintptr(unsafe.Pointer(v))) }

// U64ToPointer converts a [uint64] into a pointer of type *T.
// Used to lift a Core WebAssembly i64 into a pointer as specified in the [Canonical ABI].
//
// [uint64]: https://pkg.go.dev/builtin#uint64
// [Canonical ABI]: https://github.com/WebAssembly/component-model/blob/main/design/mvp/CanonicalABI.md
func U64ToPointer[T any](v uint64) *T { return (*T)(unsafePointer(uintptr(v))) }

// Appease vet, see https://github.com/golang/go/issues/58625
func unsafePointer(p uintptr) unsafe.Pointer {
	return *(*unsafe.Pointer)(unsafe.Pointer(&p))
}

package main

import "unsafe"

var _ unsafe.Pointer

//go:linkname C.CString runtime.cgo_CString
func C.CString(string) *C.char

//go:linkname C.GoString runtime.cgo_GoString
func C.GoString(*C.char) string

//go:linkname C.__GoStringN runtime.cgo_GoStringN
func C.__GoStringN(*C.char, uintptr) string

func C.GoStringN(cstr *C.char, length C.int) string {
	return C.__GoStringN(cstr, uintptr(length))
}

//go:linkname C.__GoBytes runtime.cgo_GoBytes
func C.__GoBytes(unsafe.Pointer, uintptr) []byte

func C.GoBytes(ptr unsafe.Pointer, length C.int) []byte {
	return C.__GoBytes(ptr, uintptr(length))
}

//go:linkname C.__CBytes runtime.cgo_CBytes
func C.__CBytes([]byte) unsafe.Pointer

func C.CBytes(b []byte) unsafe.Pointer {
	return C.__CBytes(b)
}

type (
	C.char      uint8
	C.schar     int8
	C.uchar     uint8
	C.short     int16
	C.ushort    uint16
	C.int       int32
	C.uint      uint32
	C.long      int32
	C.ulong     uint32
	C.longlong  int64
	C.ulonglong uint64
)

//export foo
func C.foo(a C.int, b C.int) C.int

var C.foo$funcaddr unsafe.Pointer

//export variadic0
//go:variadic
func C.variadic0()

var C.variadic0$funcaddr unsafe.Pointer

//export variadic2
//go:variadic
func C.variadic2(x C.int, y C.int)

var C.variadic2$funcaddr unsafe.Pointer

//export _Cgo_static_173c95a79b6df1980521_staticfunc
func C.staticfunc!symbols.go(x C.int)

var C.staticfunc!symbols.go$funcaddr unsafe.Pointer

//go:extern someValue
var C.someValue C.int

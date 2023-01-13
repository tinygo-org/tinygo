// CGo errors:
//     testdata/errors.go:4:2: warning: some warning
//     testdata/errors.go:11:9: error: unknown type name 'someType'
//     testdata/errors.go:22:5: warning: another warning
//     testdata/errors.go:13:23: unexpected token ), expected end of expression
//     testdata/errors.go:19:26: unexpected token ), expected end of expression

// Type checking errors after CGo processing:
//     testdata/errors.go:102: cannot use 2 << 10 (untyped int constant 2048) as C.char value in variable declaration (overflows)
//     testdata/errors.go:105: unknown field z in struct literal
//     testdata/errors.go:108: undefined: C.SOME_CONST_1
//     testdata/errors.go:110: cannot use C.SOME_CONST_3 (untyped int constant 1234) as byte value in variable declaration (overflows)
//     testdata/errors.go:112: undefined: C.SOME_CONST_4

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
type C._Ctype_struct___0 struct {
	x C.int
	y C.int
}
type C.point_t = C._Ctype_struct___0

const C.SOME_CONST_3 = 1234

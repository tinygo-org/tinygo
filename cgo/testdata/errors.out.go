// CGo errors:
//     testdata/errors.go:14:1: missing function name in #cgo noescape line
//     testdata/errors.go:15:1: multiple function names in #cgo noescape line
//     testdata/errors.go:4:2: warning: some warning
//     testdata/errors.go:11:9: error: unknown type name 'someType'
//     testdata/errors.go:31:5: warning: another warning
//     testdata/errors.go:18:23: unexpected token ), expected end of expression
//     testdata/errors.go:26:26: unexpected token ), expected end of expression
//     testdata/errors.go:21:33: unexpected token ), expected end of expression
//     testdata/errors.go:22:34: unexpected token ), expected end of expression
//     -: unexpected token INT, expected end of expression
//     testdata/errors.go:35:35: unexpected number of parameters: expected 2, got 3
//     testdata/errors.go:36:31: unexpected number of parameters: expected 2, got 1
//     testdata/errors.go:3:1: function "unusedFunction" in #cgo noescape line is not used

// Type checking errors after CGo processing:
//     testdata/errors.go:102: cannot use 2 << 10 (untyped int constant 2048) as _Cgo_char value in variable declaration (overflows)
//     testdata/errors.go:105: unknown field z in struct literal
//     testdata/errors.go:108: undefined: _Cgo_SOME_CONST_1
//     testdata/errors.go:110: cannot use _Cgo_SOME_CONST_3 (untyped int constant 1234) as byte value in variable declaration (overflows)
//     testdata/errors.go:112: undefined: _Cgo_SOME_CONST_4
//     testdata/errors.go:114: undefined: _Cgo_SOME_CONST_b
//     testdata/errors.go:116: undefined: _Cgo_SOME_CONST_startspace
//     testdata/errors.go:119: undefined: _Cgo_SOME_PARAM_CONST_invalid
//     testdata/errors.go:122: undefined: _Cgo_add_toomuch
//     testdata/errors.go:123: undefined: _Cgo_add_toolittle

package main

import "syscall"
import "unsafe"

var _ unsafe.Pointer

//go:linkname _Cgo_CString runtime.cgo_CString
func _Cgo_CString(string) *_Cgo_char

//go:linkname _Cgo_GoString runtime.cgo_GoString
func _Cgo_GoString(*_Cgo_char) string

//go:linkname _Cgo___GoStringN runtime.cgo_GoStringN
func _Cgo___GoStringN(*_Cgo_char, uintptr) string

func _Cgo_GoStringN(cstr *_Cgo_char, length _Cgo_int) string {
	return _Cgo___GoStringN(cstr, uintptr(length))
}

//go:linkname _Cgo___GoBytes runtime.cgo_GoBytes
func _Cgo___GoBytes(unsafe.Pointer, uintptr) []byte

func _Cgo_GoBytes(ptr unsafe.Pointer, length _Cgo_int) []byte {
	return _Cgo___GoBytes(ptr, uintptr(length))
}

//go:linkname _Cgo___CBytes runtime.cgo_CBytes
func _Cgo___CBytes([]byte) unsafe.Pointer

func _Cgo_CBytes(b []byte) unsafe.Pointer {
	return _Cgo___CBytes(b)
}

//go:linkname _Cgo___get_errno_num runtime.cgo_errno
func _Cgo___get_errno_num() uintptr

func _Cgo___get_errno() error {
	return syscall.Errno(_Cgo___get_errno_num())
}

type (
	_Cgo_char      uint8
	_Cgo_schar     int8
	_Cgo_uchar     uint8
	_Cgo_short     int16
	_Cgo_ushort    uint16
	_Cgo_int       int32
	_Cgo_uint      uint32
	_Cgo_long      int32
	_Cgo_ulong     uint32
	_Cgo_longlong  int64
	_Cgo_ulonglong uint64
)
type _Cgo_struct_point_t struct {
	x _Cgo_int
	y _Cgo_int
}
type _Cgo_point_t = _Cgo_struct_point_t

const _Cgo_SOME_CONST_3 = 1234
const _Cgo_SOME_PARAM_CONST_valid = 3 + 4

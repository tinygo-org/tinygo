// CGo errors:
//     testdata/flags.go:5:7: invalid #cgo line: NOFLAGS
//     testdata/flags.go:8:13: invalid flag: -fdoes-not-exist
//     testdata/flags.go:29:14: invalid flag: -does-not-exists

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

const C.BAR = 3
const C.FOO_H = 1

type C.int16_t = int16
type C.int32_t = int32
type C.int64_t = int64
type C.int8_t = int8
type C.uint16_t = uint16
type C.uint32_t = uint32
type C.uint64_t = uint64
type C.uint8_t = uint8
type C.uintptr_t = uintptr
type C.char uint8
type C.int int32
type C.long int32
type C.longlong int64
type C.schar int8
type C.short int16
type C.uchar uint8
type C.uint uint32
type C.ulong uint32
type C.ulonglong uint64
type C.ushort uint16

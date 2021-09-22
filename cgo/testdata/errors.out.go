// CGo errors:
//     testdata/errors.go:4:2: warning: some warning
//     testdata/errors.go:11:9: error: unknown type name 'someType'
//     testdata/errors.go:22:5: warning: another warning
//     testdata/errors.go:13:23: unexpected token ), expected end of expression
//     testdata/errors.go:19:26: unexpected token ), expected end of expression

// Type checking errors after CGo processing:
//     testdata/errors.go:102: cannot use 2 << 10 (untyped int constant 2048) as uint8 value in variable declaration (overflows)
//     testdata/errors.go:105: unknown field z in struct literal
//     testdata/errors.go:108: undeclared name: C.SOME_CONST_1
//     testdata/errors.go:110: cannot use C.SOME_CONST_3 (untyped int constant 1234) as byte value in variable declaration (overflows)
//     testdata/errors.go:112: undeclared name: C.SOME_CONST_4

package main

import "unsafe"

var _ unsafe.Pointer

const C.SOME_CONST_3 = 1234

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
type C.point_t = struct {
	x C.int
	y C.int
}

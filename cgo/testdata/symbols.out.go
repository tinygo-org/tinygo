package main

import "unsafe"

var _ unsafe.Pointer

//export foo
func C.foo(a C.int, b C.int) C.int

//export variadic0
//go:variadic
func C.variadic0()

//export variadic2
//go:variadic
func C.variadic2(x C.int, y C.int)

var C.foo$funcaddr unsafe.Pointer
var C.variadic0$funcaddr unsafe.Pointer
var C.variadic2$funcaddr unsafe.Pointer

//go:extern someValue
var C.someValue C.int

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

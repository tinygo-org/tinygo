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

const C.option2A = 20
const C.optionA = 0
const C.optionB = 1
const C.optionC = -5
const C.optionD = -4
const C.optionE = 10
const C.optionF = 11
const C.optionG = 12
const C.unused1 = 5

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
type C.bitfield_t = C.struct_4
type C.myIntArray = [10]C.int
type C.myint = C.int
type C.option2_t = C.uint
type C.option_t = C.enum_option
type C.point2d_t = struct {
	x C.int
	y C.int
}
type C.point3d_t = C.struct_point3d
type C.struct_nested_t = struct {
	begin C.point2d_t
	end   C.point2d_t
	tag   C.int

	coord C.union_2
}
type C.types_t = struct {
	f   float32
	d   float64
	ptr *C.int
}
type C.union1_t = struct{ i C.int }
type C.union2d_t = C.union_union2d
type C.union3_t = C.union_1
type C.union_nested_t = C.union_3
type C.unionarray_t = struct{ arr [10]C.uchar }

func (s *C.struct_4) bitfield_a() C.uchar { return s.__bitfield_1 & 0x1f }
func (s *C.struct_4) set_bitfield_a(value C.uchar) {
	s.__bitfield_1 = s.__bitfield_1&^0x1f | value&0x1f<<0
}
func (s *C.struct_4) bitfield_b() C.uchar {
	return s.__bitfield_1 >> 5 & 0x1
}
func (s *C.struct_4) set_bitfield_b(value C.uchar) {
	s.__bitfield_1 = s.__bitfield_1&^0x20 | value&0x1<<5
}
func (s *C.struct_4) bitfield_c() C.uchar {
	return s.__bitfield_1 >> 6
}
func (s *C.struct_4) set_bitfield_c(value C.uchar,

) { s.__bitfield_1 = s.__bitfield_1&0x3f | value<<6 }

type C.struct_4 struct {
	start        C.uchar
	__bitfield_1 C.uchar

	d C.uchar
	e C.uchar
}
type C.struct_point3d struct {
	x C.int
	y C.int
	z C.int
}
type C.struct_type1 struct {
	_type   C.int
	__type  C.int
	___type C.int
}
type C.struct_type2 struct{ _type C.int }

func (union *C.union_1) unionfield_i() *C.int   { return (*C.int)(unsafe.Pointer(&union.$union)) }
func (union *C.union_1) unionfield_d() *float64 { return (*float64)(unsafe.Pointer(&union.$union)) }
func (union *C.union_1) unionfield_s() *C.short { return (*C.short)(unsafe.Pointer(&union.$union)) }

type C.union_1 struct{ $union uint64 }

func (union *C.union_2) unionfield_area() *C.point2d_t {
	return (*C.point2d_t)(unsafe.Pointer(&union.$union))
}
func (union *C.union_2) unionfield_solid() *C.point3d_t {
	return (*C.point3d_t)(unsafe.Pointer(&union.$union))
}

type C.union_2 struct{ $union [3]uint32 }

func (union *C.union_3) unionfield_point() *C.point3d_t {
	return (*C.point3d_t)(unsafe.Pointer(&union.$union))
}
func (union *C.union_3) unionfield_array() *C.unionarray_t {
	return (*C.unionarray_t)(unsafe.Pointer(&union.$union))
}
func (union *C.union_3) unionfield_thing() *C.union3_t {
	return (*C.union3_t)(unsafe.Pointer(&union.$union))
}

type C.union_3 struct{ $union [2]uint64 }

func (union *C.union_union2d) unionfield_i() *C.int { return (*C.int)(unsafe.Pointer(&union.$union)) }
func (union *C.union_union2d) unionfield_d() *[2]float64 {
	return (*[2]float64)(unsafe.Pointer(&union.$union))
}

type C.union_union2d struct{ $union [2]uint64 }
type C.enum_option C.int
type C.enum_unused C.uint

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
type C.myint = C.int
type C.struct_point2d_t struct {
	x C.int
	y C.int
}
type C.point2d_t = C.struct_point2d_t
type C.struct_point3d struct {
	x C.int
	y C.int
	z C.int
}
type C.point3d_t = C.struct_point3d
type C.struct_type1 struct {
	_type   C.int
	__type  C.int
	___type C.int
}
type C.struct_type2 struct{ _type C.int }
type C.union_union1_t struct{ i C.int }
type C.union1_t = C.union_union1_t
type C.union_union3_t struct{ $union uint64 }

func (union *C.union_union3_t) unionfield_i() *C.int { return (*C.int)(unsafe.Pointer(&union.$union)) }
func (union *C.union_union3_t) unionfield_d() *float64 {
	return (*float64)(unsafe.Pointer(&union.$union))
}
func (union *C.union_union3_t) unionfield_s() *C.short {
	return (*C.short)(unsafe.Pointer(&union.$union))
}

type C.union3_t = C.union_union3_t
type C.union_union2d struct{ $union [2]uint64 }

func (union *C.union_union2d) unionfield_i() *C.int { return (*C.int)(unsafe.Pointer(&union.$union)) }
func (union *C.union_union2d) unionfield_d() *[2]float64 {
	return (*[2]float64)(unsafe.Pointer(&union.$union))
}

type C.union2d_t = C.union_union2d
type C.union_unionarray_t struct{ arr [10]C.uchar }
type C.unionarray_t = C.union_unionarray_t
type C._Ctype_union___0 struct{ $union [3]uint32 }

func (union *C._Ctype_union___0) unionfield_area() *C.point2d_t {
	return (*C.point2d_t)(unsafe.Pointer(&union.$union))
}
func (union *C._Ctype_union___0) unionfield_solid() *C.point3d_t {
	return (*C.point3d_t)(unsafe.Pointer(&union.$union))
}

type C.struct_struct_nested_t struct {
	begin C.point2d_t
	end   C.point2d_t
	tag   C.int

	coord C._Ctype_union___0
}
type C.struct_nested_t = C.struct_struct_nested_t
type C.union_union_nested_t struct{ $union [2]uint64 }

func (union *C.union_union_nested_t) unionfield_point() *C.point3d_t {
	return (*C.point3d_t)(unsafe.Pointer(&union.$union))
}
func (union *C.union_union_nested_t) unionfield_array() *C.unionarray_t {
	return (*C.unionarray_t)(unsafe.Pointer(&union.$union))
}
func (union *C.union_union_nested_t) unionfield_thing() *C.union3_t {
	return (*C.union3_t)(unsafe.Pointer(&union.$union))
}

type C.union_nested_t = C.union_union_nested_t
type C.enum_option = C.int
type C.option_t = C.enum_option
type C.enum_option2_t = C.uint
type C.option2_t = C.enum_option2_t
type C.struct_types_t struct {
	f   float32
	d   float64
	ptr *C.int
}
type C.types_t = C.struct_types_t
type C.myIntArray = [10]C.int
type C.struct_bitfield_t struct {
	start        C.uchar
	__bitfield_1 C.uchar

	d C.uchar
	e C.uchar
}

func (s *C.struct_bitfield_t) bitfield_a() C.uchar { return s.__bitfield_1 & 0x1f }
func (s *C.struct_bitfield_t) set_bitfield_a(value C.uchar) {
	s.__bitfield_1 = s.__bitfield_1&^0x1f | value&0x1f<<0
}
func (s *C.struct_bitfield_t) bitfield_b() C.uchar {
	return s.__bitfield_1 >> 5 & 0x1
}
func (s *C.struct_bitfield_t) set_bitfield_b(value C.uchar) {
	s.__bitfield_1 = s.__bitfield_1&^0x20 | value&0x1<<5
}
func (s *C.struct_bitfield_t) bitfield_c() C.uchar {
	return s.__bitfield_1 >> 6
}
func (s *C.struct_bitfield_t) set_bitfield_c(value C.uchar,

) { s.__bitfield_1 = s.__bitfield_1&0x3f | value<<6 }

type C.bitfield_t = C.struct_bitfield_t

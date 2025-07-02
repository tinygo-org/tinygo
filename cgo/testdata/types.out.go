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
type _Cgo_myint = _Cgo_int
type _Cgo_struct_point2d_t struct {
	x _Cgo_int
	y _Cgo_int
}
type _Cgo_point2d_t = _Cgo_struct_point2d_t
type _Cgo_struct_point3d struct {
	x _Cgo_int
	y _Cgo_int
	z _Cgo_int
}
type _Cgo_point3d_t = _Cgo_struct_point3d
type _Cgo_struct_type1 struct {
	_type   _Cgo_int
	__type  _Cgo_int
	___type _Cgo_int
}
type _Cgo_struct_type2 struct{ _type _Cgo_int }
type _Cgo_union_union1_t struct{ i _Cgo_int }
type _Cgo_union1_t = _Cgo_union_union1_t
type _Cgo_union_union3_t struct{ $union uint64 }

func (union *_Cgo_union_union3_t) unionfield_i() *_Cgo_int {
	return (*_Cgo_int)(unsafe.Pointer(&union.$union))
}
func (union *_Cgo_union_union3_t) unionfield_d() *float64 {
	return (*float64)(unsafe.Pointer(&union.$union))
}
func (union *_Cgo_union_union3_t) unionfield_s() *_Cgo_short {
	return (*_Cgo_short)(unsafe.Pointer(&union.$union))
}

type _Cgo_union3_t = _Cgo_union_union3_t
type _Cgo_union_union2d struct{ $union [2]uint64 }

func (union *_Cgo_union_union2d) unionfield_i() *_Cgo_int {
	return (*_Cgo_int)(unsafe.Pointer(&union.$union))
}
func (union *_Cgo_union_union2d) unionfield_d() *[2]float64 {
	return (*[2]float64)(unsafe.Pointer(&union.$union))
}

type _Cgo_union2d_t = _Cgo_union_union2d
type _Cgo_union_unionarray_t struct{ arr [10]_Cgo_uchar }
type _Cgo_unionarray_t = _Cgo_union_unionarray_t
type _Cgo__Ctype_union___0 struct{ $union [3]uint32 }

func (union *_Cgo__Ctype_union___0) unionfield_area() *_Cgo_point2d_t {
	return (*_Cgo_point2d_t)(unsafe.Pointer(&union.$union))
}
func (union *_Cgo__Ctype_union___0) unionfield_solid() *_Cgo_point3d_t {
	return (*_Cgo_point3d_t)(unsafe.Pointer(&union.$union))
}

type _Cgo_struct_struct_nested_t struct {
	begin _Cgo_point2d_t
	end   _Cgo_point2d_t
	tag   _Cgo_int

	coord _Cgo__Ctype_union___0
}
type _Cgo_struct_nested_t = _Cgo_struct_struct_nested_t
type _Cgo_union_union_nested_t struct{ $union [2]uint64 }

func (union *_Cgo_union_union_nested_t) unionfield_point() *_Cgo_point3d_t {
	return (*_Cgo_point3d_t)(unsafe.Pointer(&union.$union))
}
func (union *_Cgo_union_union_nested_t) unionfield_array() *_Cgo_unionarray_t {
	return (*_Cgo_unionarray_t)(unsafe.Pointer(&union.$union))
}
func (union *_Cgo_union_union_nested_t) unionfield_thing() *_Cgo_union3_t {
	return (*_Cgo_union3_t)(unsafe.Pointer(&union.$union))
}

type _Cgo_union_nested_t = _Cgo_union_union_nested_t
type _Cgo_enum_option = _Cgo_int
type _Cgo_option_t = _Cgo_enum_option
type _Cgo_enum_option2_t = _Cgo_uint
type _Cgo_option2_t = _Cgo_enum_option2_t
type _Cgo_struct_types_t struct {
	f   float32
	d   float64
	ptr *_Cgo_int
}
type _Cgo_types_t = _Cgo_struct_types_t
type _Cgo_myIntArray = [10]_Cgo_int
type _Cgo_struct_bitfield_t struct {
	start        _Cgo_uchar
	__bitfield_1 _Cgo_uchar

	d _Cgo_uchar
	e _Cgo_uchar
}

func (s *_Cgo_struct_bitfield_t) bitfield_a() _Cgo_uchar { return s.__bitfield_1 & 0x1f }
func (s *_Cgo_struct_bitfield_t) set_bitfield_a(value _Cgo_uchar) {
	s.__bitfield_1 = s.__bitfield_1&^0x1f | value&0x1f<<0
}
func (s *_Cgo_struct_bitfield_t) bitfield_b() _Cgo_uchar {
	return s.__bitfield_1 >> 5 & 0x1
}
func (s *_Cgo_struct_bitfield_t) set_bitfield_b(value _Cgo_uchar) {
	s.__bitfield_1 = s.__bitfield_1&^0x20 | value&0x1<<5
}
func (s *_Cgo_struct_bitfield_t) bitfield_c() _Cgo_uchar {
	return s.__bitfield_1 >> 6
}
func (s *_Cgo_struct_bitfield_t) set_bitfield_c(value _Cgo_uchar,

) { s.__bitfield_1 = s.__bitfield_1&0x3f | value<<6 }

type _Cgo_bitfield_t = _Cgo_struct_bitfield_t

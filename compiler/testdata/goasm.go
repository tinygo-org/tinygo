package main

import "unsafe"

func AsmSqrt(x float64) float64

func AsmAdd(x, y float64) float64

func AsmFoo(x float64) (int64, float64)

func asmExport(x float64) float64 {
	return 0
}

var asmGlobalExport int32

func AsmAllIntTypes(b bool, i int, i8 int8, i16 int16, i32 int32, i64 int64, u uint, u8 uint8, u16 uint16, u32 uint32, u64 uint64, uptr uintptr) (bool, int, int8, int16, int32, int64, uint, uint8, uint16, uint32, uint64, uintptr)

func AsmAllFloatTypes(f32 float32, f64 float64, c64 complex64, c128 complex128) (float32, float64, complex64, complex128)

// all other types, except for the func and interface types
func AsmAllOtherTypes(a [2]int, c chan int, m map[int]int, ptr *int, slice []int, str string, stru AsmStruct, uptr unsafe.Pointer) ([2]int, chan int, map[int]int, *int, []int, string, AsmStruct, unsafe.Pointer)

type AsmStruct struct {
	X, Y int
}

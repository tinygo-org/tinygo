package main

/*
#include <stdio.h>
int fortytwo(void);
#include "main.h"
#include "test.h"
int mul(int, int);
#include <string.h>
#cgo CFLAGS: -DSOME_CONSTANT=17
#define someDefine -5 + 2 * 7
bool someBool;
*/
import "C"

// int headerfunc(int a) { return a + 1; }
// static int headerfunc_static(int a) { return a - 1; }
import "C"

import "unsafe"

func main() {
	println("fortytwo:", C.fortytwo())
	println("add:", C.add(C.int(3), 5))
	var x C.myint = 3
	println("myint:", x, C.myint(5))
	println("myint size:", int(unsafe.Sizeof(x)))
	var y C.longlong = -(1 << 40)
	println("longlong:", y)
	println("global:", C.global)
	println("defined ints:", C.CONST_INT, C.CONST_INT2)
	println("defined floats:", C.CONST_FLOAT, C.CONST_FLOAT2)
	println("defined string:", C.CONST_STRING)
	println("defined char:", C.CONST_CHAR)
	println("defined expr:", C.someDefine)
	var ptr C.intPointer
	var n C.int = 15
	ptr = C.intPointer(&n)
	println("15:", *ptr)
	C.store(25, &n)
	println("25:", *ptr)
	cb := C.binop_t(C.add)
	println("callback 1:", C.doCallback(20, 30, cb))
	cb = C.binop_t(C.mul)
	println("callback 2:", C.doCallback(20, 30, cb))
	genericCallbackCall[int]()

	// variadic functions
	println("variadic0:", C.variadic0())
	println("variadic2:", C.variadic2(3, 5))

	// functions in the header C snippet
	println("headerfunc:", C.headerfunc(5))
	println("static headerfunc:", C.headerfunc_static(5))
	headerfunc_2()

	// equivalent types
	var goInt8 int8 = 5
	var _ C.int8_t = goInt8

	var _ bool = C.someBool
	var _ C._Bool = C.someBool

	// more globals
	println("bool:", C.globalBool, C.globalBool2 == true)
	println("float:", C.globalFloat)
	println("double:", C.globalDouble)
	println("complex float:", C.globalComplexFloat)
	println("complex double:", C.globalComplexDouble)
	println("complex long double:", C.globalComplexLongDouble)
	println("char match:", C.globalChar == 100)
	var voidPtr unsafe.Pointer = C.globalVoidPtrNull
	println("void* match:", voidPtr == nil, C.globalVoidPtrNull == nil, (*C.int)(C.globalVoidPtrSet) == &C.global)
	println("int64_t match:", C.globalInt64 == C.int64_t(-(2<<40)))

	// complex types
	println("struct:", C.int(unsafe.Sizeof(C.globalStruct)) == C.globalStructSize, C.globalStruct.s, C.globalStruct.l, C.globalStruct.f)
	var _ [3]C.short = C.globalArray
	println("array:", C.globalArray[0], C.globalArray[1], C.globalArray[2])
	println("union:", C.int(unsafe.Sizeof(C.globalUnion)) == C.globalUnionSize)
	C.unionSetShort(22)
	println("union s:", *C.globalUnion.unionfield_s())
	C.unionSetFloat(C.float(6.28))
	println("union f (C.float):", *C.globalUnion.unionfield_f())
	C.unionSetFloat(float32(3.14))
	println("union f (float32):", *C.globalUnion.unionfield_f())
	C.unionSetData(5, 8, 1)
	data := C.globalUnion.unionfield_data()
	println("union global data:", data[0], data[1], data[2])
	returnedUnion := printUnion(C.globalUnion)
	println("union field:", *returnedUnion.unionfield_f())
	var _ C.union_joined = C.globalUnion
	printBitfield(&C.globalBitfield)
	C.globalBitfield.set_bitfield_a(7)
	C.globalBitfield.set_bitfield_b(0)
	C.globalBitfield.set_bitfield_c(0xff)
	printBitfield(&C.globalBitfield)

	// elaborated type
	p := C.struct_point2d{x: 3, y: 5}
	println("struct:", p.x, p.y)

	// multiple anonymous structs (inside a typedef)
	var _ C.point2d_t = C.point2d_t{x: 3, y: 5}
	var _ C.point3d_t = C.point3d_t{x: 3, y: 5, z: 7}

	// nested structs/unions
	var _ C.tagged_union_t
	var _ C.nested_struct_t
	var _ C.nested_union_t

	// recursive types, test using a linked list
	list := &C.list_t{n: 3, next: &C.struct_list_t{n: 6, next: &C.list_t{n: 7, next: nil}}}
	for list != nil {
		println("n in chain:", list.n)
		list = (*C.list_t)(list.next)
	}

	// named enum
	var _ C.enum_option = C.optionA
	var _ C.option_t = C.optionA
	println("option:", C.globalOption)
	println("option A:", C.optionA)
	println("option B:", C.optionB)
	println("option C:", C.optionC)
	println("option D:", C.optionD)
	println("option E:", C.optionE)
	println("option F:", C.optionF)
	println("option G:", C.optionG)

	// anonymous enum
	var _ C.option2_t = C.option2A
	var _ C.option3_t = C.option3A
	println("option 2A:", C.option2A)
	println("option 3A:", C.option3A)

	// anonymous structs and enums in multiple Go files
	var _ C.teststruct
	var _ C.testenum

	// Check that enums are considered the same width in C and CGo.
	println("enum width matches:", unsafe.Sizeof(C.option2_t(0)) == uintptr(C.smallEnumWidth))

	// Check whether CFLAGS are correctly passed on to compiled C files.
	println("CFLAGS value:", C.cflagsConstant)

	// Check array-to-pointer decaying. This signature:
	//   void arraydecay(int buf1[5], int buf2[3][8], int buf3[4][7][2]);
	// decays to:
	//   void arraydecay(int *buf1, int *buf2[8], int *buf3[7][2]);
	C.arraydecay((*C.int)(nil), (*[8]C.int)(nil), (*[7][2]C.int)(nil))

	// Test CGo builtins like C.CString.
	cstr := C.CString("string passed to C")
	println("cstr length:", C.strlen(cstr))
	gostr := C.GoString(cstr)
	println("C.CString:", gostr)
	charBuf := C.GoBytes(unsafe.Pointer(&C.globalChars[0]), 4)
	println("C.charBuf:", charBuf[0], charBuf[1], charBuf[2], charBuf[3])
	binaryString := C.GoStringN(&C.globalChars[0], 4)
	println("C.CStringN:", len(binaryString), binaryString[0], binaryString[1], binaryString[2], binaryString[3])

	// Test whether those builtins also work with zero length data.
	println("C.GoString(nil):", C.GoString(nil))
	println("len(C.GoStringN(nil, 0)):", len(C.GoStringN(nil, 0)))
	println("len(C.GoBytes(nil, 0)):", len(C.GoBytes(nil, 0)))

	// libc: test whether C functions work at all.
	buf1 := []byte("foobar\x00")
	buf2 := make([]byte, len(buf1))
	C.strcpy((*C.char)(unsafe.Pointer(&buf2[0])), (*C.char)(unsafe.Pointer(&buf1[0])))
	println("copied string:", string(buf2[:C.strlen((*C.char)(unsafe.Pointer(&buf2[0])))]))

	// libc: test basic stdio functionality
	putsBuf := []byte("line written using C puts\x00")
	C.puts((*C.char)(unsafe.Pointer(&putsBuf[0])))

	// libc: test whether printf works in C.
	printfBuf := []byte("line written using C printf with value=%d\n\x00")
	C.printf_single_int((*C.char)(unsafe.Pointer(&printfBuf[0])), -21)
}

func printUnion(union C.joined_t) C.joined_t {
	data := union.unionfield_data()
	println("union local data: ", data[0], data[1], data[2])
	*union.unionfield_s() = -33
	println("union s:", data[0] == -33)
	*union.unionfield_f() = 6.28
	println("union f:", *union.unionfield_f())
	return union
}

//export mul
func mul(a, b C.int) C.int {
	return a * b
}

func printBitfield(bitfield *C.bitfield_t) {
	println("bitfield a:", bitfield.bitfield_a())
	println("bitfield b:", bitfield.bitfield_b())
	println("bitfield c:", bitfield.bitfield_c())
	println("bitfield d:", bitfield.d)
	println("bitfield e:", bitfield.e)
}

type Int interface {
	int | uint
}

func genericCallbackCall[T Int]() {
	println("callback inside generic function:", C.doCallback(20, 30, C.binop_t(C.add)))
}

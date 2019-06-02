package main

/*
int fortytwo(void);
#include "main.h"
int mul(int, int);
*/
import "C"

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

	// equivalent types
	var goInt8 int8 = 5
	var _ C.int8_t = goInt8

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
	println("union s:", C.globalUnion.s)
	C.unionSetFloat(3.14)
	println("union f:", C.globalUnion.f)
	C.unionSetData(5, 8, 1)
	println("union global data:", C.globalUnion.data[0], C.globalUnion.data[1], C.globalUnion.data[2])
	println("union field:", printUnion(C.globalUnion).f)
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
}

func printUnion(union C.joined_t) C.joined_t {
	println("union local data: ", union.data[0], union.data[1], union.data[2])
	union.s = -33
	println("union s:", union.data[0] == -33)
	union.f = 6.28
	println("union f:", union.f)
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

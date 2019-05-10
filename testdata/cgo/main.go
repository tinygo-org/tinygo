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

	// elaborated type
	p := C.struct_point{x: 3, y: 5}
	println("struct:", p.x, p.y)

	// recursive types, test using a linked list
	list := &C.list_t{n: 3, next: &C.struct_list_t{n: 6, next: &C.list_t{n: 7, next: nil}}}
	for list != nil {
		println("n in chain:", list.n)
		list = (*C.list_t)(list.next)
	}
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

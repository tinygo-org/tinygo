package main

/*
int fortytwo(void);
#include "main.h"
int mul(int, int);
*/
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

	// more globals
	println("bool:", C.globalBool, C.globalBool2 == true)
	println("float:", C.globalFloat)
	println("double:", C.globalDouble)
	println("complex float:", C.globalComplexFloat)
	println("complex double:", C.globalComplexDouble)
	println("complex long double:", C.globalComplexLongDouble)
	println("struct:", C.globalStruct.s, C.globalStruct.l, C.globalStruct.f)
}

//export mul
func mul(a, b C.int) C.int {
	return a * b
}

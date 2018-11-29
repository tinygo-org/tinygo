package main

/*
#include <stdint.h>
int32_t fortytwo(void);
int32_t mul(int32_t a, int32_t b);
typedef int32_t myint;
int add(int a, int b);
*/
import "C"

func main() {
	println("fortytwo:", C.fortytwo())
	println("mul:", C.mul(C.int32_t(3), 5))
	println("add:", C.add(C.int(3), 5))
	var x C.myint = 3
	println("myint:", x, C.myint(5))
	var y C.longlong = -(1 << 40)
	println("longlong:", y)
}

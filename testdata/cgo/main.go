package main

/*
#include <stdint.h>
int32_t fortytwo(void);
int32_t mul(int32_t a, int32_t b);
*/
import "C"

func main() {
	println("fortytwo:", C.fortytwo())
	println("mul:", C.mul(int32(3), 5))
}

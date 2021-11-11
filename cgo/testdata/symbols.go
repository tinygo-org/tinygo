package main

/*
// Function signatures.
int foo(int a, int b);
void variadic0();
void variadic2(int x, int y, ...);

// Global variable signatures.
extern int someValue;
*/
import "C"

// Test function signatures.
func accessFunctions() {
	C.foo(3, 4)
	C.variadic0()
	C.variadic2(3, 5)
}

func accessGlobals() {
	_ = C.someValue
}

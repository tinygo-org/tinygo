package main

/*
// Function signatures.
int foo(int a, int b);
void variadic0();
void variadic2(int x, int y, ...);
static void staticfunc(int x);

// Global variable signatures.
extern int someValue;

void notEscapingFunction(int *a);

#cgo noescape notEscapingFunction
*/
import "C"

// Test function signatures.
func accessFunctions() {
	C.foo(3, 4)
	C.variadic0()
	C.variadic2(3, 5)
	C.staticfunc(3)
	C.notEscapingFunction(nil)
}

func accessGlobals() {
	_ = C.someValue
}

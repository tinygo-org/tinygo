package main

/*
// Function signatures.
int foo(int a, int b);
void variadic0();
void variadic2(int x, int y, ...);
void variadicNoescape(int *a, ...);
static void staticfunc(int x);
void arrayParam(int a[4]);
void incompleteArrayParam(const int a[]);

// Global variable signatures.
extern int someValue;

void notEscapingFunction(int *a);

#cgo noescape notEscapingFunction
#cgo noescape variadicNoescape
*/
import "C"

// Test function signatures.
func accessFunctions() {
	C.foo(3, 4)
	C.variadic0()
	C.variadic2(3, 5)
	C.variadicNoescape(nil)
	C.staticfunc(3)
	C.arrayParam(nil)
	C.incompleteArrayParam(nil)
	C.notEscapingFunction(nil)
}

func accessGlobals() {
	_ = C.someValue
}

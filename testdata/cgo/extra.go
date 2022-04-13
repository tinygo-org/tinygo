package main

// Make sure CGo supports multiple files.

// int fortytwo(void);
// static float headerfunc_static(float a) { return a - 1; }
// static void headerfunc_void(int a, int *ptr) { *ptr = a; }
import "C"

func headerfunc_2() {
	// Call headerfunc_static that is different from the headerfunc_static in
	// the main.go file.
	// The upstream CGo implementation does not handle this case correctly.
	println("static headerfunc 2:", C.headerfunc_static(5))

	// Test function without return value.
	var n C.int
	C.headerfunc_void(3, &n)
	println("static headerfunc void:", n)
}

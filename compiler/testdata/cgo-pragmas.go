package main

/*
void variadicNoescape(int *a, ...);
#cgo noescape variadicNoescape
*/
import "C"

func callVariadicNoescape(a *C.int) {
	C.variadicNoescape(a)
}

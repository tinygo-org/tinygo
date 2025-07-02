package main

/*
#define foo 3
#define bar foo

#define unreferenced 4
#define referenced unreferenced

#define fnlike() 5
#define fnlike_val fnlike()
#define square(n) (n*n)
#define square_val square(20)
#define add(a, b) (a + b)
#define add_val add(3, 5)
*/
import "C"

const (
	Foo = C.foo
	Bar = C.bar

	Baz = C.referenced

	fnlike = C.fnlike_val
	square = C.square_val
	add    = C.add_val
)

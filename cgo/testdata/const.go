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

typedef unsigned char flag_t;
#define cast_val ((flag_t)(1 << 0))
#define cast_fn(x) ((flag_t)x)
#define cast_fn_val cast_fn(2)
#define cast_int ((int)-1)
#define cast_float ((float)1)
*/
import "C"

const (
	Foo = C.foo
	Bar = C.bar

	Baz = C.referenced

	fnlike = C.fnlike_val
	square = C.square_val
	add    = C.add_val

	castVal   = C.cast_val
	castFnVal = C.cast_fn_val
	castInt   = C.cast_int
	castFloat = C.cast_float
)

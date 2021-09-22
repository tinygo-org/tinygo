package main

/*
#warning some warning

typedef struct {
	int x;
	int y;
} point_t;

typedef someType noType; // undefined type

#define SOME_CONST_1 5) // invalid const syntax
#define SOME_CONST_2 6) // const not used (so no error)
#define SOME_CONST_3 1234 // const too large for byte
*/
//
//
// #define SOME_CONST_4 8) // after some empty lines
import "C"

// #warning another warning
import "C"

// Make sure that errors for the following lines won't change with future
// additions to the CGo preamble.
//line errors.go:100
var (
	// constant too large
	_ C.uint8_t = 2 << 10

	// z member does not exist
	_ C.point_t = C.point_t{z: 3}

	// constant has syntax error
	_ = C.SOME_CONST_1

	_ byte = C.SOME_CONST_3

	_ = C.SOME_CONST_4
)

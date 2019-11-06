package main

/*
// Simple typedef.
typedef int myint;

// Structs, with or without name.
typedef struct {
	int x;
	int y;
} point2d_t;
typedef struct point3d {
	int x;
	int y;
	int z;
} point3d_t;

// Structs with reserved field names.
struct type1 {
	// All these fields should be renamed.
	int type;
	int _type;
	int __type;
};
struct type2 {
	// This field should not be renamed.
	int _type;
};

// Enums. These define constant numbers. All these constants must be given the
// correct number.
typedef enum option {
	optionA,
	optionB,
	optionC = -5,
	optionD,
	optionE = 10,
	optionF,
	optionG,
} option_t;

// Anonymous enum.
typedef enum {
	option2A = 20,
} option2_t;

// Various types that are usually translated directly to Go types, but storing
// them in a struct reveals them.
typedef struct {
	float  f;
	double d;
	int    *ptr;
} types_t;

// Arrays.
typedef int myIntArray[10];

// Bitfields.
typedef struct {
	unsigned char start;
	unsigned char a : 5;
	unsigned char b : 1;
	unsigned char c : 2;
	unsigned char :0; // new field
	unsigned char d : 6;
	unsigned char e : 3;
	// Note that C++ allows bitfields bigger than the underlying type.
} bitfield_t;
*/
import "C"

var (
	// Simple typedefs.
	_ C.myint

	// Structs.
	_ C.point2d_t
	_ C.point3d_t
	_ C.struct_point3d

	// Structs with reserved field names.
	_ C.struct_type1
	_ C.struct_type2

	// Enums (anonymous and named).
	_ C.option_t
	_ C.enum_option
	_ C.option2_t

	// Various types.
	_ C.types_t

	// Arrays.
	_ C.myIntArray
	_ C.myIntArrayPtr
)

// Test bitfield accesses.
func foo() {
	var x C.bitfield_t
	x.start = 3
	x.a = 4
	x.b = 1
	x.c = 2
	x.d = 10
	x.e = 5
}

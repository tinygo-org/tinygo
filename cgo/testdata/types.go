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

// Unions.
typedef union {
	// Union should be treated as a struct.
	int i;
} union1_t;
typedef union {
	// Union must contain a single field and have special getters/setters.
	int    i;
	double d;
	short  s;
} union3_t;
typedef union union2d {
	int i;
	double d[2];
} union2d_t;
typedef union {
	unsigned char arr[10];
} unionarray_t;

// Nested structs and unions.
typedef struct {
	point2d_t begin;
	point2d_t end;
	int       tag;
	union {
		point2d_t area;
		point3d_t solid;
	} coord;
} struct_nested_t;
typedef union {
	point3d_t    point;
	unionarray_t array;
	union3_t     thing;
} union_nested_t;

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
enum unused {
	unused1 = 5,
};

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

// Function signatures.
void variadic0();
void variadic2(int x, int y, ...);
*/
import "C"

// // Test that we can refer from this CGo fragment to the fragment above.
// typedef myint myint2;
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

	// Unions.
	_ C.union1_t
	_ C.union3_t
	_ C.union2d_t
	_ C.unionarray_t

	// Nested structs and unions.
	_ C.struct_nested_t
	_ C.union_nested_t

	// Enums (anonymous and named).
	_ C.option_t
	_ C.enum_option
	_ C.option2_t

	// Various types.
	_ C.types_t

	// Arrays.
	_ C.myIntArray
)

// Test bitfield accesses.
func accessBitfields() {
	var x C.bitfield_t
	x.start = 3
	x.set_bitfield_a(4)
	x.set_bitfield_b(1)
	x.set_bitfield_c(2)
	x.d = 10
	x.e = 5
	var _ C.uchar = x.bitfield_a()
}

// Test union accesses.
func accessUnion() {
	var union1 C.union1_t
	union1.i = 5

	var union2d C.union2d_t
	var _ *C.int = union2d.unionfield_i()
	var _ *[2]float64 = union2d.unionfield_d()
}

// Test function signatures.
func accessFunctions() {
	C.variadic0()
	C.variadic2(3, 5)
}

#include <stdbool.h>
#include <stdint.h>

typedef short myint;
typedef short unusedTypedef;
int add(int a, int b);
int unusedFunction(void);
typedef int (*binop_t) (int, int);
int doCallback(int a, int b, binop_t cb);
typedef int * intPointer;
void store(int value, int *ptr);

int variadic0();
int variadic2(int x, int y, ...);

# define CONST_INT 5
# define CONST_INT2 5llu
# define CONST_FLOAT 5.8
# define CONST_FLOAT2 5.8f
# define CONST_CHAR 'c'
# define CONST_STRING "defined string"

// this signature should not be included by CGo
void unusedFunction2(int x, __builtin_va_list args);

typedef struct collection {
	short         s;
	long          l;
	float         f;
	unsigned char c;
} collection_t;

struct point2d {
	int x;
	int y;
};

typedef struct {
	int x;
	int y;
} point2d_t;

typedef struct {
	int x;
	int y;
	int z;
} point3d_t;

typedef struct {
	int tag;
	union {
		int a;
		int b;
	} u;
} tagged_union_t;

typedef struct {
	int x;
	struct {
		char red;
		char green;
		char blue;
	} color;
} nested_struct_t;

typedef union {
	int x;
	struct {
		char a;
		char b;
	};
} nested_union_t;

// linked list
typedef struct list_t {
	int           n;
	struct list_t *next;
} list_t;

typedef union joined {
	myint s;
	float f;
	short data[3];
} joined_t;
void unionSetShort(short s);
void unionSetFloat(float f);
void unionSetData(short f0, short f1, short f2);

typedef enum option {
	optionA,
	optionB,
	optionC = -5,
	optionD,
	optionE = 10,
	optionF,
	optionG,
} option_t;

typedef enum {
	option2A = 20,
} option2_t;

typedef enum {
	option3A = 21,
} option3_t;

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

// test globals and datatypes
extern int global;
extern int unusedGlobal;
extern bool globalBool;
extern bool globalBool2;
extern float globalFloat;
extern double globalDouble;
extern _Complex float globalComplexFloat;
extern _Complex double globalComplexDouble;
extern _Complex double globalComplexLongDouble;
extern char globalChar;
extern void *globalVoidPtrSet;
extern void *globalVoidPtrNull;
extern int64_t globalInt64;
extern collection_t globalStruct;
extern int globalStructSize;
extern short globalArray[3];
extern joined_t globalUnion;
extern int globalUnionSize;
extern option_t globalOption;
extern bitfield_t globalBitfield;

extern int smallEnumWidth;

extern int cflagsConstant;

extern char globalChars[4];

// test duplicate definitions
int add(int a, int b);
extern int global;

// Test array decaying into a pointer.
typedef int arraydecay_buf3[4][7][2];
void arraydecay(int buf1[5], int buf2[3][8], arraydecay_buf3 buf3);

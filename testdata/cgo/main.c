#include <math.h>
#include "main.h"
#include <stdio.h>

int global = 3;
bool globalBool = 1;
bool globalBool2 = 10; // test narrowing
float globalFloat = 3.1;
double globalDouble = 3.2;
_Complex float globalComplexFloat = 4.1+3.3i;
_Complex double globalComplexDouble = 4.2+3.4i;
_Complex double globalComplexLongDouble = 4.3+3.5i;
char globalChar = 100;
void *globalVoidPtrSet = &global;
void *globalVoidPtrNull;
int64_t globalInt64 = -(2LL << 40);
collection_t globalStruct = {256, -123456, 3.14, 88};
int globalStructSize = sizeof(globalStruct);
short globalArray[3] = {5, 6, 7};
joined_t globalUnion;
int globalUnionSize = sizeof(globalUnion);
option_t globalOption = optionG;
bitfield_t globalBitfield = {244, 15, 1, 2, 47, 5};

int cflagsConstant = SOME_CONSTANT;

int smallEnumWidth = sizeof(option2_t);

char globalChars[] = {2, 0, 4, 8};

int fortytwo() {
	return 42;
}

int add(int a, int b) {
	return a + b;
}

int doCallback(int a, int b, binop_t callback) {
	return callback(a, b);
}

int variadic0() {
	return 1;
}

int variadic2(int x, int y, ...) {
	return x * y;
}

void store(int value, int *ptr) {
	*ptr = value;
}

void unionSetShort(short s) {
	globalUnion.s = s;
}

void unionSetFloat(float f) {
	globalUnion.f = f;
}

void unionSetData(short f0, short f1, short f2) {
	globalUnion.data[0] = 5;
	globalUnion.data[1] = 8;
	globalUnion.data[2] = 1;
}

void arraydecay(int buf1[5], int buf2[3][8], int buf3[4][7][2]) {
	// Do nothing.
}

double doSqrt(double x) {
	return sqrt(x);
}

void printf_single_int(char *format, int arg) {
	printf(format, arg);
}

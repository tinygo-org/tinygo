#include "main.h"

int global = 3;
_Bool globalBool = 1;
_Bool globalBool2 = 10; // test narrowing
float globalFloat = 3.1;
double globalDouble = 3.2;
_Complex float globalComplexFloat = 4.1+3.3i;
_Complex double globalComplexDouble = 4.2+3.4i;
_Complex double globalComplexLongDouble = 4.3+3.5i;
char globalChar = 100;
collection_t globalStruct = {256, -123456, 3.14, 88};
int globalStructSize = sizeof(globalStruct);
short globalArray[3] = {5, 6, 7};
joined_t globalUnion;
int globalUnionSize = sizeof(globalUnion);

int fortytwo() {
	return 42;
}

int add(int a, int b) {
	return a + b;
}

int doCallback(int a, int b, binop_t callback) {
	return callback(a, b);
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

#include <math.h>
#include "main.h"
#include <stdio.h>
#include <stdlib.h>

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

int set_errno(int err) {
	errno = err;
	return -1;
}

typedef struct malloc_node {
	struct malloc_node *next;
	int value;
} malloc_node;

void *makeMallocChain(void) {
	malloc_node *tail = malloc(sizeof(malloc_node));
	tail->next = NULL;
	tail->value = 42;

	malloc_node *head = malloc(sizeof(malloc_node));
	head->next = tail;
	head->value = 1;
	return head;
}

void clobberMalloc(void) {
#if defined(__AVR__)
	return;
#else
	malloc_node *nodes[64];
	for (int i = 0; i < 64; i++) {
		nodes[i] = malloc(sizeof(malloc_node));
		nodes[i]->next = NULL;
		nodes[i]->value = 0;
	}
	for (int i = 0; i < 64; i++) {
		free(nodes[i]);
	}
#endif
}

int mallocChainValue(void *ptr) {
	return ((malloc_node *)ptr)->next->value;
}

#define MALLOC_HIDE_MASK ((uintptr_t)0x5a5a5a5a)

__attribute__((noinline)) uintptr_t makeHiddenMalloc(void) {
	malloc_node *node = malloc(sizeof(malloc_node));
	node->next = NULL;
	node->value = 84;
	return (uintptr_t)node ^ MALLOC_HIDE_MASK;
}

int hiddenMallocValue(uintptr_t hidden) {
	malloc_node *node = (malloc_node *)(hidden ^ MALLOC_HIDE_MASK);
	return node->value;
}

void freeHiddenMalloc(uintptr_t hidden) {
	free((void *)(hidden ^ MALLOC_HIDE_MASK));
}

void mallocFreeStress(void) {
#if defined(__AVR__)
	const int count = 32;
#else
	const int count = 1024;
#endif
	for (int i = 0; i < count; i++) {
		char *ptr = malloc(1024);
		ptr[0] = (char)i;
		free(ptr);
	}
}

void mallocZero(void) {
	free(malloc(0));
}

__attribute__((noinline)) void *callCalloc(size_t nmemb, size_t size) {
	return calloc(nmemb, size);
}

int callocOverflowReturnsNull(void) {
#if defined(__linux__) || defined(_WIN32) || defined(__APPLE__)
	return 1;
#else
	volatile size_t nmemb = (size_t)-1;
	return callCalloc(nmemb, 2) == NULL;
#endif
}

__attribute__((noinline)) void clobberStack(void) {
#if defined(__AVR__)
	return;
#else
	volatile uintptr_t values[128];
	for (int i = 0; i < 128; i++) {
		values[i] = 0;
	}
#endif
}

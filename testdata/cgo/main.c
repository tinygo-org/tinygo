#include "main.h"

int global = 3;

int fortytwo() {
	return 42;
}

int add(int a, int b) {
	return a + b;
}

void store(int value, int *ptr) {
	*ptr = value;
}

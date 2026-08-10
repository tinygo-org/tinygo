#include <stdlib.h>

int reallocPreservesContents(void) {
	int *ptr = malloc(sizeof(int));
	*ptr = 42;
	ptr = realloc(ptr, 2 * sizeof(int));
	int value = ptr[0];
	free(ptr);
	return value;
}

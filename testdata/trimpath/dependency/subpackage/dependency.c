#include "shared.h"
#include "relative.h"

_Static_assert(RELATIVE_VALUE == 43, "incorrect relative include");

int sharedValue(void) {
	return sharedValueInline();
}

const char *sharedHeaderPath(void) {
	return sharedHeaderPathInline();
}

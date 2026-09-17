#include "shared.h"

int sharedValue(void) {
	return sharedValueInline();
}

const char *sharedHeaderPath(void) {
	return sharedHeaderPathInline();
}

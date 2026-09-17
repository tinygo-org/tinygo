#include "relative.h"
#include <stdint.h>

_Static_assert(RELATIVE_VALUE == 42, "incorrect relative include");
int32_t value(void) { return RELATIVE_VALUE; }

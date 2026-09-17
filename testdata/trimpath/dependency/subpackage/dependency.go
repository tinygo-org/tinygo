package dependency

/*
#cgo CFLAGS: -I../include -Iinclude
#include "relative.h"
_Static_assert(RELATIVE_VALUE == 43, "incorrect relative include");
int sharedValue(void);
const char *sharedHeaderPath(void);
*/
import "C"

//go:noinline
func Value() int { return int(C.sharedValue()) }

//go:noinline
func HeaderPath() string { return C.GoString(C.sharedHeaderPath()) }

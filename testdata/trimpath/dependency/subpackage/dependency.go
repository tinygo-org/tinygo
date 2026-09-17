package dependency

/*
#cgo CFLAGS: -I../include
int sharedValue(void);
const char *sharedHeaderPath(void);
*/
import "C"

//go:noinline
func Value() int { return int(C.sharedValue()) }

//go:noinline
func HeaderPath() string { return C.GoString(C.sharedHeaderPath()) }

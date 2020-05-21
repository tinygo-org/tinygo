package main

/*
// this name doesn't exist
#cgo  NOFLAGS: -foo

// unknown flag
#cgo CFLAGS: -fdoes-not-exist -DNOTDEFINED

#cgo CFLAGS: -DFOO

#cgo CFLAGS: -Iinclude
#include "foo.h"

#if defined(FOO)
#define BAR 3
#else
#define BAR 5
#endif

#if defined(NOTDEFINED)
#warning flag must not be defined
#endif

// Check Compiler flags
#cgo LDFLAGS: -lc

// This flag is not valid ldflags
#cgo LDFLAGS: -does-not-exists

*/
import "C"

var (
	_ = C.BAR
	_ = C.FOO_H
)

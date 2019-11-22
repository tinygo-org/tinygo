package main

/*
// this name doesn't exist
#cgo  NOFLAGS: -foo

// unknown flag
#cgo CFLAGS: -fdoes-not-exist -DNOTDEFINED

#cgo CFLAGS: -DFOO

#if defined(FOO)
#define BAR 3
#else
#define BAR 5
#endif

#if defined(NOTDEFINED)
#warning flag must not be defined
#endif
*/
import "C"

var (
	_ = C.BAR
)

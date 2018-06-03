
// +build linux

package runtime

import (
	"unsafe"
)

// #include <stdio.h>
// #include <stdlib.h>
// #include <unistd.h>
import "C"

const Microsecond = 1

func putchar(c byte) {
	C.putchar(C.int(c))
}

func Sleep(d Duration) {
	C.usleep(C.useconds_t(d))
}

func abort() {
	C.abort()
}

func alloc(size uintptr) unsafe.Pointer {
	buf := C.calloc(1, C.size_t(size))
	if buf == nil {
		panic("cannot allocate memory")
	}
	return buf
}

func free(ptr unsafe.Pointer) {
	C.free(ptr)
}

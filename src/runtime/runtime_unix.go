
// +build linux

package runtime

import (
	"unsafe"
)

// #include <stdio.h>
// #include <stdlib.h>
// #include <unistd.h>
// #include <time.h>
import "C"

const Microsecond = 1

func putchar(c byte) {
	C.putchar(C.int(c))
}

func sleep(d Duration) {
	C.usleep(C.useconds_t(d))
}

// Return monotonic time in microseconds.
//
// TODO: use nanoseconds?
// TODO: noescape
func monotime() uint64 {
	var ts C.struct_timespec
	C.clock_gettime(C.CLOCK_MONOTONIC, &ts)
	return uint64(ts.tv_sec) * 1000 * 1000 + uint64(ts.tv_nsec) / 1000
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
	//C.free(ptr) // TODO
}

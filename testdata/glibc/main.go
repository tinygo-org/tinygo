package main

// This program is built with -libc=glibc. It calls into the system glibc, which
// is linked dynamically, so it only runs if the binary names the dynamic linker
// and if glibc owns the thread pointer.

/*
#include <errno.h>
#include <stdlib.h>
#include <string.h>
#include <gnu/libc-version.h>

// rangeErrno returns the errno that strtol sets for a value that does not fit
// in a long. errno is a thread-local variable in glibc, so this reads glibc's
// TLS through the thread pointer.
static int rangeErrno(void) {
	errno = 0;
	strtol("99999999999999999999999999", NULL, 10);
	return errno;
}
*/
import "C"

import (
	"runtime"
	"strings"
	"unsafe"
)

func main() {
	// gnu_get_libc_version exists only in glibc, so this does not link with
	// musl.
	version := C.GoString(C.gnu_get_libc_version())
	println("glibc version 2.x:", strings.HasPrefix(version, "2."))

	// malloc and free use the tcache, which glibc keeps in TLS.
	cs := C.CString("hello from glibc")
	dup := C.strdup(cs)
	println("strdup:", C.GoString(dup))
	C.free(unsafe.Pointer(dup))
	C.free(unsafe.Pointer(cs))

	println("errno is ERANGE:", C.rangeErrno() == C.ERANGE)

	// Do the same on a goroutine, which has its own thread under the threads
	// scheduler, and thus its own glibc TLS block.
	done := make(chan bool)
	go func() {
		p := C.malloc(64)
		C.free(p)
		done <- C.rangeErrno() == C.ERANGE
	}()
	println("errno is ERANGE on goroutine:", <-done)

	runtime.GC()
	println("done")
}

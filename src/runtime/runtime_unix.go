
// +build linux

package runtime

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

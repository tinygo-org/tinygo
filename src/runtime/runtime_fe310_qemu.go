// +build fe310,qemu

package runtime

import (
	"runtime/volatile"
	"unsafe"
)

const tickMicros = 100 // CLINT.MTIME increments every 100ns

// Special memory-mapped device to exit tests, created by SiFive.
var testExit = (*volatile.Register32)(unsafe.Pointer(uintptr(0x100000)))

var timestamp timeUnit

func abort() {
	// Signal a successful exit.
	testExit.Set(0x5555)
}

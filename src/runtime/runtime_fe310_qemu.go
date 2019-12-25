// +build fe310,qemu

package runtime

import (
	"runtime/volatile"
	"unsafe"
)

// Special memory-mapped device to exit tests, created by SiFive.
var testExit = (*volatile.Register32)(unsafe.Pointer(uintptr(0x100000)))

var timestamp timeUnit

func abort() {
	// Signal a successful exit.
	testExit.Set(0x5555)
}

func ticks() timeUnit {
	return timestamp
}

func sleepTicks(d timeUnit) {
	// Note: QEMU doesn't seem to support the RTC peripheral at the time of
	// writing so just simulate sleeping here.
	timestamp += d
}

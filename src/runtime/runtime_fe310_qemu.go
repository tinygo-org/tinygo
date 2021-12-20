// +build fe310,qemu

package runtime

import (
	"runtime/volatile"
	"unsafe"
)

// Special memory-mapped device to exit tests, created by SiFive.
var testExit = (*volatile.Register32)(unsafe.Pointer(uintptr(0x100000)))

// ticksToNanoseconds converts CLINT ticks (at 100ns per tick) to nanoseconds.
func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 100
}

// nanosecondsToTicks converts nanoseconds to CLINT ticks (at 100ns per tick).
func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns / 100)
}

func abort() {
	exit(1)
}

func exit(code int) {
	if code == 0 {
		// Signal a successful exit.
		testExit.Set(0x5555) // FINISHER_PASS
	} else {
		// Signal a failure. The exit code is stored in the upper 16 bits of the
		// 32 bit value.
		testExit.Set(uint32(code)<<16 | 0x3333) // FINISHER_FAIL
	}
}

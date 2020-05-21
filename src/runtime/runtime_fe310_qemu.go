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
	// Signal a successful exit.
	testExit.Set(0x5555)
}

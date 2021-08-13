// +build k210,!qemu

package runtime

import (
	"device/riscv"
)

// ticksToNanoseconds converts CPU ticks to nanoseconds.
func ticksToNanoseconds(ticks timeUnit) int64 {
	// The following calculation is actually the following, but with both sides
	// reduced to reduce the risk of overflow:
	//     ticks * 1e9 / (390000000 / 50)
	// 50 is the CLINT divider and 390000000 is the CPU frequency.
	return int64(ticks) * 5000 / 39
}

// nanosecondsToTicks converts nanoseconds to CPU ticks.
func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns * 39 / 5000)
}

func exit(code int) {
	abort()
}

func abort() {
	// lock up forever
	for {
		riscv.Asm("wfi")
	}
}

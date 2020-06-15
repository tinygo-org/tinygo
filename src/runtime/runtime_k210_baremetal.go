// +build k210,!qemu

package runtime

import (
	"device/kendryte"
	"device/riscv"
)

var clockFrequency uint32 = kendryte.SYSCTL.CLK_FREQ.Get()

// ticksToNanoseconds converts RTC ticks to nanoseconds.
func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks) * 1e9 / clockFrequency
}

// nanosecondsToTicks converts nanoseconds to RTC ticks.
func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns * 64 / 1953125)
}

func abort() {
	// lock up forever
	for {
		riscv.Asm("wfi")
	}
}

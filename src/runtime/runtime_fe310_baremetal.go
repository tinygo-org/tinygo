// +build fe310,!qemu

package runtime

import (
	"device/riscv"
	"device/sifive"
)

func abort() {
	// lock up forever
	for {
		riscv.Asm("wfi")
	}
}

func ticks() timeUnit {
	// Combining the low bits and the high bits yields a time span of over 270
	// years without counter rollover.
	highBits := sifive.RTC.RTCHI.Get()
	for {
		lowBits := sifive.RTC.RTCLO.Get()
		newHighBits := sifive.RTC.RTCHI.Get()
		if newHighBits == highBits {
			// High bits stayed the same.
			println("bits:", highBits, lowBits)
			return timeUnit(lowBits) | (timeUnit(highBits) << 32)
		}
		// Retry, because there was a rollover in the low bits (happening every
		// 1.5 days).
		highBits = newHighBits
	}
}

func sleepTicks(d timeUnit) {
	target := ticks() + d
	for ticks() < target {
	}
}

// +build fe310,!qemu

package runtime

import (
	"device/riscv"
)

const tickMicros = 32768 // RTC clock runs at 32.768kHz

func abort() {
	// lock up forever
	for {
		riscv.Asm("wfi")
	}
}

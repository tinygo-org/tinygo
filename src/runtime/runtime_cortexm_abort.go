// +build cortexm,!nxp

package runtime

import (
	"device/arm"
)

func abort() {
	// disable all interrupts
	arm.DisableInterrupts()

	// lock up forever
	for {
		arm.Asm("wfi")
	}
}

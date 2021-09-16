// +build cortexm,!nxp,!qemu

package runtime

import (
	"device/arm"
)

var ResetOnAbort = false

func exit(code int) {
	abort()
}

func abort() {
	if ResetOnAbort {
		arm.SystemReset()
	}
	// lock up forever
	for {
		arm.Asm("wfi")
	}
}

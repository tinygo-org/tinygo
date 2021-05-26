// +build cortexm,!nxp,!qemu

package runtime

import (
	"device/arm"
)

func abort() {
	// lock up forever
	for {
		arm.Asm("wfi")
	}
}

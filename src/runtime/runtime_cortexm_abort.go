// +build cortexm,!nxp,!qemu

package runtime

import (
	"github.com/sago35/device/arm"
)

func abort() {
	// lock up forever
	for {
		arm.Asm("wfi")
	}
}

//go:build cortexm && !nxp && !qemu

package runtime

import (
	"device/arm"
)

func exit(code int) {
	abort()
}

func abort() {
	// lock up forever
	for {
		arm.Asm("wfi")
	}
}

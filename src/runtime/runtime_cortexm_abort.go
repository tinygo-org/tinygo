//go:build cortexm && !nxp && !qemu
// +build cortexm,!nxp,!qemu

package runtime

import (
	"tinygo.org/x/device/arm"
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

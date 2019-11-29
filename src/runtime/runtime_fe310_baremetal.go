// +build fe310,!qemu

package runtime

import (
	"device/riscv"
)

func abort() {
	// lock up forever
	for {
		riscv.Asm("wfi")
	}
}

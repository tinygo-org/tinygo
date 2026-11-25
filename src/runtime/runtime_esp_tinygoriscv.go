//go:build esp && tinygo.riscv

package runtime

import "device/riscv"

func exit(code int) {
	abort()
}

func abort() {
	for {
		riscv.Asm("wfi")
	}
}

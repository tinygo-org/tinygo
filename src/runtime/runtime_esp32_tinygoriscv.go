//go:build esp32c3 || esp32c6 || esp32h2 || esp32p4

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

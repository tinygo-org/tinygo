//go:build esp32 && qemu

package runtime

import "device"

func exit(code int) {
	qemuExit(code)
}

func abort() {
	qemuExit(1)
}

func qemuExit(code int) {
	// QEMU semihosting expects a2 = SYS_exit (1) and a3 = the exit code.
	// AsmFull cannot declare these clobbers, so no Go code may run afterward.
	device.AsmFull(
		"mov a3, {code}\n"+
			"movi a2, 1\n"+
			"simcall",
		map[string]interface{}{"code": code},
	)
	// This loop ensures the clobbered registers are never used again.
	for {
		device.Asm("waiti 0")
	}
}

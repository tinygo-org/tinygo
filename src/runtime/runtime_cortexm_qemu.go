//go:build cortexm && qemu

package runtime

// This file implements the Stellaris LM3S6965 Cortex-M3 chip as implemented by
// QEMU.

import (
	"device/arm"
	"runtime/volatile"
	"unsafe"
)

//export Reset_Handler
func main() {
	preinit()
	initSysTick()
	run()

	// Signal successful exit.
	exit(0)
}

// QEMU clocks the LM3S6965 at 200 MHz / (SYSDIV + 1) = 12.5 MHz from its reset RCC.
// See ssys_calculate_system_clock in QEMU hw/arm/stellaris.c.
func sysTickFrequency() uint32 {
	return 12_500_000
}

// UART0 output register.
var stdoutWrite = (*volatile.Register8)(unsafe.Pointer(uintptr(0x4000c000)))

func putchar(c byte) {
	stdoutWrite.Set(uint8(c))
}

func getchar() byte {
	// dummy, TODO
	return 0
}

func buffered() int {
	// dummy, TODO
	return 0
}

func abort() {
	exit(1)
}

func exit(code int) {
	// Exit QEMU.
	if code == 0 {
		arm.SemihostingCall(arm.SemihostingReportException, arm.SemihostingApplicationExit)
	} else {
		arm.SemihostingCall(arm.SemihostingReportException, arm.SemihostingRunTimeErrorUnknown)
	}

	// Lock up forever (should be unreachable).
	for {
		arm.Asm("wfi")
	}
}

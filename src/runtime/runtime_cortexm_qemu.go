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

// UART0 data (UARTDR) and flag (UARTFR) registers.
// See https://github.com/qemu/qemu/blob/master/hw/char/pl011.c
var (
	uartData = (*volatile.Register8)(unsafe.Pointer(uintptr(0x4000c000)))
	uartFlag = (*volatile.Register32)(unsafe.Pointer(uintptr(0x4000c018)))
)

// uartReceived reports whether a byte was received, which UARTFR shows by
// clearing bit 4 (RXFE).
func uartReceived() bool {
	return uartFlag.Get()&(1<<4) == 0
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

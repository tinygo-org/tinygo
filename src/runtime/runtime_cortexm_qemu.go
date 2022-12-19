//go:build cortexm && qemu

package runtime

// This file implements the Stellaris LM3S6965 Cortex-M3 chip as implemented by
// QEMU.

import (
	"device/arm"
	"runtime/volatile"
	"unsafe"
)

type timeUnit int64

var timestamp timeUnit

//export Reset_Handler
func main() {
	preinit()
	run()

	// Signal successful exit.
	exit(0)
}

func ticksToNanoseconds(ticks timeUnit) int64 {
	return int64(ticks)
}

func nanosecondsToTicks(ns int64) timeUnit {
	return timeUnit(ns)
}

func sleepTicks(d timeUnit) {
	// TODO: actually sleep here for the given time.
	timestamp += d
}

func ticks() timeUnit {
	return timestamp
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

func waitForEvents() {
	arm.Asm("wfe")
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

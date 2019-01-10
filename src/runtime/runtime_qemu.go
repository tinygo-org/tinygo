// +build qemu

package runtime

// This file implements the Stellaris LM3S6965 Cortex-M3 chip as implemented by
// QEMU.

import (
	"device/arm"
	"unsafe"
)

type timeUnit int64

const tickMicros = 1

var timestamp timeUnit

//go:export Reset_Handler
func main() {
	preinit()
	initAll()
	callMain()
	arm.SemihostingCall(arm.SemihostingReportException, arm.SemihostingApplicationExit)
	abort()
}

const asyncScheduler = false

func sleepTicks(d timeUnit) {
	// TODO: actually sleep here for the given time.
	timestamp += d
}

func ticks() timeUnit {
	return timestamp
}

//go:volatile
type regValue uint32

// UART0 output register.
var stdoutWrite *regValue = (*regValue)(unsafe.Pointer(uintptr(0x4000c000)))

func putchar(c byte) {
	*stdoutWrite = regValue(c)
}

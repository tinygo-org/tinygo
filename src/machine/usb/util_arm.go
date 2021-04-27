// +build arm

package usb

import "device/arm"

//go:linkname ticks runtime.ticks
func ticks() int64

// udelay waits for the given number of microseconds before returning.
// We cannot use the sleep timer from this context (import cycle), but we need
// an approximate method to spin CPU cycles for short periods of time.
//go:inline
func udelay(microsec uint32) {
	n := cycles(microsec, descCPUFrequencyHz)
	for i := uint32(0); i < n; i++ {
		arm.Asm(`nop`)
	}
}

func disableInterrupts() uintptr {
	return arm.DisableInterrupts()
}

func enableInterrupts(mask uintptr) {
	arm.EnableInterrupts(mask)
}

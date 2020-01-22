// +build nxp

package machine

import (
	"device/arm"
	"device/nxp"
)

type PinMode uint8

const (
	PinInput PinMode = iota
	PinOutput
)

// Stop enters STOP (deep sleep) mode
func Stop() {
	// set SLEEPDEEP to enable deep sleep
	nxp.SystemControl.SCR.SetBits(nxp.SystemControl_SCR_SLEEPDEEP)

	// enter STOP mode
	arm.Asm("wfi")
}

// Wait enters WAIT (sleep) mode
func Wait() {
	// clear SLEEPDEEP bit to disable deep sleep
	nxp.SystemControl.SCR.ClearBits(nxp.SystemControl_SCR_SLEEPDEEP)

	// enter WAIT mode
	arm.Asm("wfi")
}

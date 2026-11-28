//go:build rp2040 || rp2350

package machine

import (
	"device/rp"
	"runtime/interrupt"
)

// CPUReset performs a hard system reset using the watchdog.
//
// SYSRESETREQ resets only the processors and leaves USB in a bad state.
// This follows watchdog_reboot in pico-sdk hardware_watchdog/watchdog.c.
func CPUReset() {
	interrupt.Disable()
	rp.WATCHDOG.CTRL.ClearBits(rp.WATCHDOG_CTRL_ENABLE)

	rp.PSM.WDSEL.Set(watchdogResetMask)

	// Clear the bootrom watchdog boot magic so it does a normal boot.
	rp.WATCHDOG.SCRATCH4.Set(0)

	rp.WATCHDOG.CTRL.SetBits(rp.WATCHDOG_CTRL_TRIGGER)
	for {
	}
}

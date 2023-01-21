//go:build rp2040

package machine

import (
	"device/arm"
	"device/rp"
)

// Sleep until RTC interrupt
func Sleep() {

	scb_orig := arm.SCB.SCR.Get()
	clock0_orig := clocks.sleepEN0.Get()
	clock1_orig := clocks.sleepEN1.Get()

	// Stop all clocks but RTC
	clocks.sleep()

	// Turn off all clocks when in sleep mode except for RTC
	clocks.sleepEN0.Set(rp.CLOCKS_SLEEP_EN0_CLK_RTC_RTC)
	clocks.sleepEN1.Set(0x0)

	// Enable deep sleep at the proc
	arm.SCB.SCR.SetBits(arm.SCB_SCR_SLEEPDEEP)

	// Go to sleep
	arm.Asm("wfi")

	// Restore state
	arm.SCB.SCR.Set(scb_orig)
	clocks.sleepEN0.Set(clock0_orig)
	clocks.sleepEN1.Set(clock1_orig)

	// Enable clocks
	clocks.init()
}

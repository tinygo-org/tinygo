//go:build nrf && softdevice
// +build nrf,softdevice

package runtime

import (
	"tinygo.org/x/device/arm"
	"tinygo.org/x/device/nrf"
)

//export sd_app_evt_wait
func sd_app_evt_wait()

// This is a global variable to avoid a heap allocation in waitForEvents.
var softdeviceEnabled uint8

func waitForEvents() {
	// Call into the SoftDevice to sleep. This is necessary here because a
	// normal wfe will not put the chip in low power mode (it still consumes
	// 500µA-1mA). It is really needed to call sd_app_evt_wait for low power
	// consumption.

	// First check whether the SoftDevice is enabled. Unfortunately,
	// sd_app_evt_wait cannot be called when the SoftDevice is not enabled.
	arm.SVCall1(0x12, &softdeviceEnabled) // sd_softdevice_is_enabled

	if softdeviceEnabled != 0 {
		// Now pick the appropriate SVCall number. Hopefully they won't change
		// in the future with a different SoftDevice version.
		if nrf.Device == "nrf51" {
			// sd_app_evt_wait: SOC_SVC_BASE_NOT_AVAILABLE + 29
			arm.SVCall0(0x2B + 29)
		} else if nrf.Device == "nrf52" || nrf.Device == "nrf52840" || nrf.Device == "nrf52833" {
			// sd_app_evt_wait: SOC_SVC_BASE_NOT_AVAILABLE + 21
			arm.SVCall0(0x2C + 21)
		} else {
			sd_app_evt_wait()
		}
	} else {
		// SoftDevice is disabled so we can sleep normally.
		arm.Asm("wfe")
	}
}

//go:build nrf && !softdevice
// +build nrf,!softdevice

package runtime

import "device/arm"

func waitForEvents() {
	arm.Asm("wfe")
}

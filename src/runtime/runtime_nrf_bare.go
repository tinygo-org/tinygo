//go:build nrf && !softdevice
// +build nrf,!softdevice

package runtime

import "tinygo.org/x/device/arm"

func waitForEvents() {
	arm.Asm("wfe")
}

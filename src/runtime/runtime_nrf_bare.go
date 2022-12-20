//go:build nrf && !softdevice

package runtime

import "device/arm"

func waitForEvents() {
	arm.Asm("wfe")
}

// +build nrf,!softdevice

package runtime

import "github.com/sago35/device/arm"

func waitForEvents() {
	arm.Asm("wfe")
}

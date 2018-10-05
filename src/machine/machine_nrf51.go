// +build nrf51

package machine

import (
	"device/nrf"
)

// Get peripheral and pin number for this GPIO pin.
func (p GPIO) getPortPin() (*nrf.GPIO_Type, uint8) {
	return nrf.GPIO, p.Pin
}

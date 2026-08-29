//go:build py32 && py32_no_gpio_afrh

package machine

import "device/py32"

// SetAltFunc selects a pin-specific alternate function.
// Configure the pin as PinAlternate first; AF values are listed in the device
// datasheet and vary by pin and device.
func (p Pin) SetAltFunc(af uint8) {
	port, pin := p.getPort()
	port.AFRL.ReplaceBits(uint32(af), py32.GPIO_AFRL_AFSEL0_Msk, (pin%8)*4)
}

//go:build py32 && no_gpio_afrh

package machine

import "device/py32"

// SetAltFunc selects the alternate function for a GPIO pin on PY32F devices.
//
// Each pin supports up to 16 alternate functions (AF0–AF15), encoded as a
// 4-bit field in the GPIO alternate-function registers. The register is split
// in two: AFRL holds the 4-bit selectors for pins 0–7 and AFRH holds them for
// pins 8–15, with each selector at bit offset (pin%8)*4 within its register.
//
// Alternate-function mappings vary by pin and PY32 variant; consult the
// device datasheet for the AF0-AF15 mapping. The pin must also be configured
// as PinAlternate via Pin.Configure before the alternate function takes effect.
func (p Pin) SetAltFunc(af uint8) {
	port, pin := p.getPort()
	port.AFRL.ReplaceBits(uint32(af), py32.GPIO_AFRL_AFSEL0_Msk, (pin%8)*4)
}

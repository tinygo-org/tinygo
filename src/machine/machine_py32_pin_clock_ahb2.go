//go:build py32 && py32_gpio_clock_ahb2

package machine

import "device/py32"

func (p Pin) enableClock() {
	py32.RCC.AHB2ENR.SetBits(py32.RCC_AHB2ENR_IOPAEN << p.getPortNumber())
}

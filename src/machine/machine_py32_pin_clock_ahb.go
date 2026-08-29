//go:build py32 && py32_gpio_clock_ahb

package machine

import "device/py32"

func (p Pin) enableClock() {
	py32.RCC.AHBENR.SetBits(py32.RCC_AHBENR_IOPAEN << p.getPortNumber())
}

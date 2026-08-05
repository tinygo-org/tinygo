//go:build py32 && !py32_gpio_clock_ahb1 && !py32_gpio_clock_ahb

package machine

import "device/py32"

func (p Pin) enableClock() {
	py32.RCC.IOPENR.SetBits(1 << p.getPortNumber())
}

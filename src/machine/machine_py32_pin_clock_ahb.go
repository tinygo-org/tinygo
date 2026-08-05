//go:build py32 && py32_gpio_clock_ahb

package machine

import "device/py32"

func (p Pin) enableClock() {
	// PY32F410 places IOPAEN at bit 8 of AHBENR.
	py32.RCC.AHBENR.SetBits(1 << (p.getPortNumber() + 8))
}

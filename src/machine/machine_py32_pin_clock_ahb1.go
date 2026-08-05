//go:build py32 && py32_gpio_clock_ahb1

package machine

import "device/py32"

func (p Pin) enableClock() {
	// PY32E407/F403 use STM32F4-style GPIO clock bits starting at bit 0.
	py32.RCC.AHB1ENR.SetBits(1 << p.getPortNumber())
}

//go:build rp2350

package machine

import (
	"device/rp"
	"unsafe"
)

const (
	_NUMBANK0_GPIOS = 48
	_NUMBANK0_IRQS  = 6
	xoscFreq        = 12 // Pico 2 Crystal oscillator Abracon ABM8-272-T3 frequency in MHz
	rp2350ExtraReg  = 1
	notimpl         = "rp2350: not implemented"
	initUnreset     = rp.RESETS_RESET_ADC |
		rp.RESETS_RESET_SPI0 |
		rp.RESETS_RESET_SPI1 |
		rp.RESETS_RESET_UART0 |
		rp.RESETS_RESET_UART1 |
		rp.RESETS_RESET_USBCTRL
)

var (
	timer = (*timerType)(unsafe.Pointer(rp.TIMER0))
)

// Enable or disable a specific interrupt on the executing core.
// num is the interrupt number which must be in [0,31].
func irqSet(num uint32, enabled bool) {
	if num >= _NUMIRQ {
		return
	}
	irqSetMask(1<<num, enabled)
}

func irqSetMask(mask uint32, enabled bool) {
	panic(notimpl)
}

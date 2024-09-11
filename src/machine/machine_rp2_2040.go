//go:build rp2040

package machine

import (
	"device/rp"
	"unsafe"
)

const (
	_NUMBANK0_GPIOS = 30
	_NUMBANK0_IRQS  = 4
	rp2350ExtraReg  = 0
	initUnreset     = rp.RESETS_RESET_ADC |
		rp.RESETS_RESET_RTC |
		rp.RESETS_RESET_SPI0 |
		rp.RESETS_RESET_SPI1 |
		rp.RESETS_RESET_UART0 |
		rp.RESETS_RESET_UART1 |
		rp.RESETS_RESET_USBCTRL
)

var (
	timer = (*timerType)(unsafe.Pointer(rp.TIMER))
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
	if enabled {
		// Clear pending before enable
		// (if IRQ is actually asserted, it will immediately re-pend)
		rp.PPB.NVIC_ICPR.Set(mask)
		rp.PPB.NVIC_ISER.Set(mask)
	} else {
		rp.PPB.NVIC_ICER.Set(mask)
	}
}

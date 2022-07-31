//go:build rp2040
// +build rp2040

package machine

import (
	"device/rp"
)

const deviceName = rp.Device

const (
	// GPIO pins
	GPIO0  Pin = 0
	GPIO1  Pin = 1
	GPIO2  Pin = 2
	GPIO3  Pin = 3
	GPIO4  Pin = 4
	GPIO5  Pin = 5
	GPIO6  Pin = 6
	GPIO7  Pin = 7
	GPIO8  Pin = 8
	GPIO9  Pin = 9
	GPIO10 Pin = 10
	GPIO11 Pin = 11
	GPIO12 Pin = 12
	GPIO13 Pin = 13
	GPIO14 Pin = 14
	GPIO15 Pin = 15
	GPIO16 Pin = 16
	GPIO17 Pin = 17
	GPIO18 Pin = 18
	GPIO19 Pin = 19
	GPIO20 Pin = 20
	GPIO21 Pin = 21
	GPIO22 Pin = 22
	GPIO23 Pin = 23
	GPIO24 Pin = 24
	GPIO25 Pin = 25
	GPIO26 Pin = 26
	GPIO27 Pin = 27
	GPIO28 Pin = 28
	GPIO29 Pin = 29

	// Analog pins
	ADC0 Pin = GPIO26
	ADC1 Pin = GPIO27
	ADC2 Pin = GPIO28
	ADC3 Pin = GPIO29
)

//go:linkname machineInit runtime.machineInit
func machineInit() {
	// Reset all peripherals to put system into a known state,
	// except for QSPI pads and the XIP IO bank, as this is fatal if running from flash
	// and the PLLs, as this is fatal if clock muxing has not been reset on this boot
	// and USB, syscfg, as this disturbs USB-to-SWD on core 1
	bits := ^uint32(rp.RESETS_RESET_IO_QSPI |
		rp.RESETS_RESET_PADS_QSPI |
		rp.RESETS_RESET_PLL_USB |
		rp.RESETS_RESET_USBCTRL |
		rp.RESETS_RESET_SYSCFG |
		rp.RESETS_RESET_PLL_SYS)
	resetBlock(bits)

	// Remove reset from peripherals which are clocked only by clkSys and
	// clkRef. Other peripherals stay in reset until we've configured clocks.
	bits = ^uint32(rp.RESETS_RESET_ADC |
		rp.RESETS_RESET_RTC |
		rp.RESETS_RESET_SPI0 |
		rp.RESETS_RESET_SPI1 |
		rp.RESETS_RESET_UART0 |
		rp.RESETS_RESET_UART1 |
		rp.RESETS_RESET_USBCTRL)
	unresetBlockWait(bits)

	clocks.init()

	// Peripheral clocks should now all be running
	unresetBlockWait(RESETS_RESET_Msk)
}

//go:linkname ticks runtime.machineTicks
func ticks() uint64 {
	return timer.timeElapsed()
}

//go:linkname lightSleep runtime.machineLightSleep
func lightSleep(ticks uint64) {
	timer.lightSleep(ticks)
}

// CurrentCore returns the core number the call was made from.
func CurrentCore() int {
	return int(rp.SIO.CPUID.Get())
}

// NumCores returns number of cores available on the device.
func NumCores() int { return 2 }

func Sleep(pin Pin, change PinChange) error {
	if pin > 31 {
		return ErrInvalidInputPin
	}

	base := &ioBank0.dormantWakeIRQctrl
	pin.ctrlSetInterrupt(change, true, base) // gpio_set_dormant_irq_enabled

	clocks.deinit() // Reconfigure clocks, disable PLLs etc

	xosc.Dormant() // Execution stops here until woken up

	clocks.init() // Re-initialize clocks

	// Clear the irq so we can go back to dormant mode again if we want
	pin.acknowledgeInterrupt(change)
	return nil
}

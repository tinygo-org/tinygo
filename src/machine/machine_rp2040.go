//go:build rp2040
// +build rp2040

package machine

import (
	"device/rp"
)

const deviceName = rp.Device

const (
	// GPIO pins
	GPIO0  Pin = 0  // peripherals: PWM0 channel A
	GPIO1  Pin = 1  // peripherals: PWM0 channel B
	GPIO2  Pin = 2  // peripherals: PWM1 channel A
	GPIO3  Pin = 3  // peripherals: PWM1 channel B
	GPIO4  Pin = 4  // peripherals: PWM2 channel A
	GPIO5  Pin = 5  // peripherals: PWM2 channel B
	GPIO6  Pin = 6  // peripherals: PWM3 channel A
	GPIO7  Pin = 7  // peripherals: PWM3 channel B
	GPIO8  Pin = 8  // peripherals: PWM4 channel A
	GPIO9  Pin = 9  // peripherals: PWM4 channel B
	GPIO10 Pin = 10 // peripherals: PWM5 channel A
	GPIO11 Pin = 11 // peripherals: PWM5 channel B
	GPIO12 Pin = 12 // peripherals: PWM6 channel A
	GPIO13 Pin = 13 // peripherals: PWM6 channel B
	GPIO14 Pin = 14 // peripherals: PWM7 channel A
	GPIO15 Pin = 15 // peripherals: PWM7 channel B
	GPIO16 Pin = 16 // peripherals: PWM0 channel A
	GPIO17 Pin = 17 // peripherals: PWM0 channel B
	GPIO18 Pin = 18 // peripherals: PWM1 channel A
	GPIO19 Pin = 19 // peripherals: PWM1 channel B
	GPIO20 Pin = 20 // peripherals: PWM2 channel A
	GPIO21 Pin = 21 // peripherals: PWM2 channel B
	GPIO22 Pin = 22 // peripherals: PWM3 channel A
	GPIO23 Pin = 23 // peripherals: PWM3 channel B
	GPIO24 Pin = 24 // peripherals: PWM4 channel A
	GPIO25 Pin = 25 // peripherals: PWM4 channel B
	GPIO26 Pin = 26 // peripherals: PWM5 channel A
	GPIO27 Pin = 27 // peripherals: PWM5 channel B
	GPIO28 Pin = 28 // peripherals: PWM6 channel A
	GPIO29 Pin = 29 // peripherals: PWM6 channel B

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

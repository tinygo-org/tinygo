// +build stm32

package machine

import "unsafe"

// Peripheral abstraction layer for the stm32.

type PinMode uint8

// AltFunc represents the alternate function peripherals that can be mapped to
// the GPIO ports. Since these differ by what is supported on the stm32 family
// they are defined in the more specific files
type AltFunc uint8

// Peripheral operations sequence:
//  1. Enable the clock to the alternate function.
//  2. Enable clock to corresponding GPIO
//  3. Attach the alternate function.
//  4. Configure the input-output port and pins (of the corresponding GPIOx) to match the AF .
//  5. If desired enable the nested vector interrupt control to generate interrupts.
//  6. Program the AF/peripheral for the required configuration (eg baud rate for a USART) .

// Given that the stm32 family has the AF and GPIO on different registers based on the chip,
//  use the main function here for configuring, and use hooks in the more specific chip
//  definition files
// Also, the stm32f1xx series handles things differently from the stm32f0/2/3/4

// EnableClock configures the STM32 to send clock signals to the given GPIO port
func (p Pin) EnableClock() {
	p.enableClock()
}

// Implement these in the more specialized family files, extended to include the ports
//  available for that device
/*
// getPort is a helper to map to the I/O ports. Extend to the
//  supported ports on your part.
func (p Pin) getPort() *stm32.GPIO_Type {
	switch p / 16 {
	case 0:
		return stm32.GPIOA
	case 1:
		return stm32.GPIOB
	case 2:
		...
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
// Use the appropriate APB1ENR/APB2ENR for the part
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPAEN)
	case 1:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPBEN)
	case 2:
		...
	default:
		panic("machine: unknown port")
	}
}

// configure sets the port I/O direction, speed, pull-up/down config etc
func (p Pin) configure(config PinConfig) {
	// perform GPIO configuration specific to this family
}
*/

// ---------- General pin operations ----------

// Configure this pin with the given I/O settings.
func (p Pin) Configure(config PinConfig) {
	p.EnableClock()
	p.configure(config)
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p Pin) Set(high bool) {
	port := p.getPort()
	pin := uint8(p) % 16
	if high {
		port.BSRR.Set(1 << pin)
	} else {
		port.BSRR.Set(1 << (pin + 16))
	}
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	port := p.getPort()
	pin := uint8(p) % 16
	val := port.IDR.Get() & (1 << pin)
	return (val > 0)
}

// EnableAltFuncClock sends clock signals to the alternative function peripheral
func EnableAltFuncClock(bus unsafe.Pointer) {
	enableAltFuncClock(bus)
}

//go:build stm32

package machine

import "device/stm32"

const deviceName = stm32.Device

// Peripheral abstraction layer for the stm32.

const (
	portA Pin = iota * 16
	portB
	portC
	portD
	portE
	portF
	portG
	portH
	portI
	portJ
	portK
)

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

// ---------- General pin operations ----------
type PinChange uint8

const (
	PinRising PinChange = 1 << iota
	PinFalling
	PinToggle = PinRising | PinFalling
)

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

// Get returns the current value of a GPIO pin when the pin is configured as an
// input or as an output.
func (p Pin) Get() bool {
	port := p.getPort()
	pin := uint8(p) % 16
	val := port.IDR.Get() & (1 << pin)
	return (val > 0)
}

// PortMaskSet returns the register and mask to enable a given GPIO pin. This
// can be used to implement bit-banged drivers.
func (p Pin) PortMaskSet() (*uint32, uint32) {
	port := p.getPort()
	pin := uint8(p) % 16
	return &port.BSRR.Reg, 1 << pin
}

// PortMaskClear returns the register and mask to disable a given port. This can
// be used to implement bit-banged drivers.
func (p Pin) PortMaskClear() (*uint32, uint32) {
	port := p.getPort()
	pin := uint8(p) % 16
	return &port.BSRR.Reg, 1 << (pin + 16)
}

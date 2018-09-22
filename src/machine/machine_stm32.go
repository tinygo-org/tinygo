// +build stm32

package machine

// Peripheral abstraction layer for the stm32.

import (
	"device/stm32"
)

type GPIOMode uint8

const (
	GPIO_INPUT  = 0 // Input mode
	GPIO_OUTPUT = 2 // Output mode, max speed 2MHz
)

const (
	portA = iota * 16
	portB
	portC
	portD
	portE
	portF
	portG
)

func (p GPIO) getPort() *stm32.GPIO_Type {
	switch p.Pin / 16 {
	case 0:
		return stm32.GPIOA
	case 1:
		return stm32.GPIOB
	case 2:
		return stm32.GPIOC
	case 3:
		return stm32.GPIOD
	case 4:
		return stm32.GPIOE
	case 5:
		return stm32.GPIOF
	case 6:
		return stm32.GPIOG
	default:
		panic("machine: unknown port")
	}
}

// Configure this pin with the given configuration.
func (p GPIO) Configure(config GPIOConfig) {
	// Enable clock.
	// Do this always, as it isn't known whether the clock has already been
	// enabled.
	stm32.RCC.APB2ENR |= stm32.RCC_APB2ENR_IOPCEN_Msk;

	// Configure the GPIO pin.
	port := p.getPort()
	pin := p.Pin % 16
	pos := p.Pin % 8 * 4
	if pin < 8 {
		port.CRL = stm32.RegValue((uint32(port.CRL) &^ (0xf << pos)) | (uint32(config.Mode) << pos))
	} else {
		port.CRH = stm32.RegValue((uint32(port.CRH) &^ (0xf << pos)) | (uint32(config.Mode) << pos))
	}
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p GPIO) Set(high bool) {
	port := p.getPort()
	pin := p.Pin % 16
	if high {
		port.BSRR = 1 << pin
	} else {
		port.BSRR = 1 << (pin + 16)
	}
}

// +build stm32,!stm32f103xx

package machine

import (
	"device/stm32"
)

// GPIO for the stm32 families except the stm32f1xx which uses a simpler but
//  less flexible mechanism. Extend the +build directive above to exclude other
//  models in the stm32f1xx series as necessary

const (
	// Mode Flag
	PinOutput        PinMode = 0
	PinInput         PinMode = PinInputFloating
	PinInputFloating PinMode = 1
	PinInputPulldown PinMode = 2
	PinInputPullup   PinMode = 3

	// for UART
	PinModeUARTTX PinMode = 4
	PinModeUARTRX PinMode = 5

	// for I2C
	PinModeI2CSCL PinMode = 6
	PinModeI2CSDA PinMode = 7

	// for SPI
	PinModeSPICLK  PinMode = 8
	PinModeSPIMOSI PinMode = 9
	PinModeSPIMISO PinMode = 10

	// for analog/ADC
	PinInputAnalog PinMode = 11

	// for PWM
	// TBD
)

// Configure this pin with the given configuration
func (p Pin) Configure(config PinConfig) {
	// Use the default system alternate function; this
	//  will only be used if you try to call this with
	//  one of the peripheral modes instead of vanilla GPIO.
	p.ConfigureAltFunc(config, stm32.AF0_SYSTEM)
}

// Configure this pin with the given configuration including alternate
//  function mapping if necessary.
func (p Pin) ConfigureAltFunc(config PinConfig, altFunc stm32.AltFunc) {
	// Configure the GPIO pin.
	p.enableClock()
	port := p.getPort()
	pin := uint8(p) % 16
	pos := pin * 2

	switch config.Mode {

	// GPIO
	case PinInputFloating:
		port.MODER.ReplaceBits(stm32.GPIOModeInput, 0x3, pos)
		port.PUPDR.ReplaceBits(stm32.GPIOPUPDRFloating, 0x3, pos)
	case PinInputPulldown:
		port.MODER.ReplaceBits(stm32.GPIOModeInput, 0x3, pos)
		port.PUPDR.ReplaceBits(stm32.GPIOPUPDRPullDown, 0x3, pos)
	case PinInputPullup:
		port.MODER.ReplaceBits(stm32.GPIOModeInput, 0x3, pos)
		port.PUPDR.ReplaceBits(stm32.GPIOPUPDRPullUp, 0x3, pos)
	case PinOutput:
		port.MODER.ReplaceBits(stm32.GPIOModeOutputGeneral, 0x3, pos)
		port.OSPEEDR.ReplaceBits(stm32.GPIOSpeedHigh, 0x3, pos)

	// UART
	case PinModeUARTTX:
		port.MODER.ReplaceBits(stm32.GPIOModeOutputAltFunc, 0x3, pos)
		port.OSPEEDR.ReplaceBits(stm32.GPIOSpeedHigh, 0x3, pos)
		port.PUPDR.ReplaceBits(stm32.GPIOPUPDRPullUp, 0x3, pos)
		p.SetAltFunc(altFunc)
	case PinModeUARTRX:
		port.MODER.ReplaceBits(stm32.GPIOModeOutputAltFunc, 0x3, pos)
		port.PUPDR.ReplaceBits(stm32.GPIOPUPDRFloating, 0x3, pos)
		p.SetAltFunc(altFunc)

	// I2C)
	case PinModeI2CSCL:
		port.MODER.ReplaceBits(stm32.GPIOModeOutputAltFunc, 0x3, pos)
		port.OTYPER.ReplaceBits(stm32.GPIOOutputTypeOpenDrain, 0x1, pos)
		port.OSPEEDR.ReplaceBits(stm32.GPIOSpeedLow, 0x3, pos)
		port.PUPDR.ReplaceBits(stm32.GPIOPUPDRFloating, 0x3, pos)
		p.SetAltFunc(altFunc)
	case PinModeI2CSDA:
		port.MODER.ReplaceBits(stm32.GPIOModeOutputAltFunc, 0x3, pos)
		port.OTYPER.ReplaceBits(stm32.GPIOOutputTypeOpenDrain, 0x1, pos)
		port.OSPEEDR.ReplaceBits(stm32.GPIOSpeedLow, 0x3, pos)
		port.PUPDR.ReplaceBits(stm32.GPIOPUPDRFloating, 0x3, pos)
		p.SetAltFunc(altFunc)

	// SPI
	case PinModeSPICLK:
		port.MODER.ReplaceBits(stm32.GPIOModeOutputAltFunc, 0x3, pos)
		port.OSPEEDR.ReplaceBits(stm32.GPIOSpeedLow, 0x3, pos)
		port.PUPDR.ReplaceBits(stm32.GPIOPUPDRFloating, 0x3, pos)
		p.SetAltFunc(altFunc)
	case PinModeSPIMOSI:
		port.MODER.ReplaceBits(stm32.GPIOModeOutputAltFunc, 0x3, pos)
		port.OSPEEDR.ReplaceBits(stm32.GPIOSpeedLow, 0x3, pos)
		port.PUPDR.ReplaceBits(stm32.GPIOPUPDRFloating, 0x3, pos)
		p.SetAltFunc(altFunc)
	case PinModeSPIMISO:
		port.MODER.ReplaceBits(stm32.GPIOModeOutputAltFunc, 0x3, pos)
		port.OSPEEDR.ReplaceBits(stm32.GPIOSpeedLow, 0x3, pos)
		port.PUPDR.ReplaceBits(stm32.GPIOPUPDRFloating, 0x3, pos)
		p.SetAltFunc(altFunc)
	}
}

// SetAltFunc maps the given alternative function to the I/O pin
func (p Pin) SetAltFunc(af stm32.AltFunc) {
	port := p.getPort()
	pin := uint8(p) % 16
	pos := (pin % 8) * 4
	if pin < 8 {
		port.AFRL.ReplaceBits(uint32(af), 0xf, pos)
	} else {
		port.AFRH.ReplaceBits(uint32(af), 0xf, pos)
	}
}

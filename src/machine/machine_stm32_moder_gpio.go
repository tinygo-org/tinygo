//go:build stm32 && !stm32f103
// +build stm32,!stm32f103

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
	PinModeSPICLK PinMode = 8
	PinModeSPISDO PinMode = 9
	PinModeSPISDI PinMode = 10

	// for analog/ADC
	PinInputAnalog PinMode = 11

	// for PWM
	PinModePWMOutput PinMode = 12
)

// Define several bitfields that have different names across chip families but
// essentially have the same meaning.
const (
	// MODER bitfields.
	gpioModeInput     = 0
	gpioModeOutput    = 1
	gpioModeAlternate = 2
	gpioModeAnalog    = 3
	gpioModeMask      = 0x3

	// PUPDR bitfields.
	gpioPullFloating = 0
	gpioPullUp       = 1
	gpioPullDown     = 2
	gpioPullMask     = 0x3

	// OSPEED bitfields.
	gpioOutputSpeedVeryHigh = 3
	gpioOutputSpeedHigh     = 2
	gpioOutputSpeedMedium   = 1
	gpioOutputSpeedLow      = 0
	gpioOutputSpeedMask     = 0x3
)

// Configure this pin with the given configuration
func (p Pin) Configure(config PinConfig) {
	// Use the default system alternate function; this
	//  will only be used if you try to call this with
	//  one of the peripheral modes instead of vanilla GPIO.
	p.ConfigureAltFunc(config, 0)
}

// Configure this pin with the given configuration including alternate
//  function mapping if necessary.
func (p Pin) ConfigureAltFunc(config PinConfig, altFunc uint8) {
	// Configure the GPIO pin.
	p.enableClock()
	port := p.getPort()
	pos := (uint8(p) % 16) * 2 // assume each field is two bits in size (with mask 0x3)

	switch config.Mode {

	// GPIO
	case PinInputFloating:
		port.MODER.ReplaceBits(gpioModeInput, gpioModeMask, pos)
		port.PUPDR.ReplaceBits(gpioPullFloating, gpioPullMask, pos)
	case PinInputPulldown:
		port.MODER.ReplaceBits(gpioModeInput, gpioModeMask, pos)
		port.PUPDR.ReplaceBits(gpioPullDown, gpioPullMask, pos)
	case PinInputPullup:
		port.MODER.ReplaceBits(gpioModeInput, gpioModeMask, pos)
		port.PUPDR.ReplaceBits(gpioPullUp, gpioPullMask, pos)
	case PinOutput:
		port.MODER.ReplaceBits(gpioModeOutput, gpioModeMask, pos)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedHigh, gpioOutputSpeedMask, pos)

	// UART
	case PinModeUARTTX:
		port.MODER.ReplaceBits(gpioModeAlternate, gpioModeMask, pos)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedHigh, gpioOutputSpeedMask, pos)
		port.PUPDR.ReplaceBits(gpioPullUp, gpioPullMask, pos)
		p.SetAltFunc(altFunc)
	case PinModeUARTRX:
		port.MODER.ReplaceBits(gpioModeAlternate, gpioModeMask, pos)
		port.PUPDR.ReplaceBits(gpioPullFloating, gpioPullMask, pos)
		p.SetAltFunc(altFunc)

	// I2C
	case PinModeI2CSCL:
		port.MODER.ReplaceBits(gpioModeAlternate, gpioModeMask, pos)
		port.OTYPER.ReplaceBits(stm32.GPIO_OTYPER_OT0_OpenDrain, stm32.GPIO_OTYPER_OT0_Msk, pos/2)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedLow, gpioOutputSpeedMask, pos)
		port.PUPDR.ReplaceBits(gpioPullUp, gpioPullMask, pos)
		p.SetAltFunc(altFunc)
	case PinModeI2CSDA:
		port.MODER.ReplaceBits(gpioModeAlternate, gpioModeMask, pos)
		port.OTYPER.ReplaceBits(stm32.GPIO_OTYPER_OT0_OpenDrain, stm32.GPIO_OTYPER_OT0_Msk, pos/2)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedLow, gpioOutputSpeedMask, pos)
		port.PUPDR.ReplaceBits(gpioPullUp, gpioPullMask, pos)
		p.SetAltFunc(altFunc)

	// SPI
	case PinModeSPICLK:
		port.MODER.ReplaceBits(gpioModeAlternate, gpioModeMask, pos)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedHigh, gpioOutputSpeedMask, pos)
		port.PUPDR.ReplaceBits(gpioPullFloating, gpioPullMask, pos)
		p.SetAltFunc(altFunc)
	case PinModeSPISDO:
		port.MODER.ReplaceBits(gpioModeAlternate, gpioModeMask, pos)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedLow, gpioOutputSpeedMask, pos)
		port.PUPDR.ReplaceBits(gpioPullFloating, gpioPullMask, pos)
		p.SetAltFunc(altFunc)
	case PinModeSPISDI:
		port.MODER.ReplaceBits(gpioModeAlternate, gpioModeMask, pos)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedLow, gpioOutputSpeedMask, pos)
		port.PUPDR.ReplaceBits(gpioPullFloating, gpioPullMask, pos)
		p.SetAltFunc(altFunc)

	// PWM
	case PinModePWMOutput:
		port.MODER.ReplaceBits(gpioModeAlternate, gpioModeMask, pos)
		port.OSPEEDR.ReplaceBits(gpioOutputSpeedHigh, gpioOutputSpeedMask, pos)
		port.PUPDR.ReplaceBits(gpioPullFloating, gpioPullMask, pos)
		p.SetAltFunc(altFunc)

	// ADC
	case PinInputAnalog:
		port.MODER.ReplaceBits(gpioModeAnalog, gpioModeMask, pos)
	}
}

// SetAltFunc maps the given alternative function to the I/O pin
func (p Pin) SetAltFunc(af uint8) {
	port := p.getPort()
	pin := uint8(p) % 16
	pos := (pin % 8) * 4
	if pin < 8 {
		port.AFRL.ReplaceBits(uint32(af), 0xf, pos)
	} else {
		port.AFRH.ReplaceBits(uint32(af), 0xf, pos)
	}
}

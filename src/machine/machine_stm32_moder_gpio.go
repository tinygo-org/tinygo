// +build stm32,!stm32f103xx

package machine

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

	// Register values for the chip
	// GPIOx_MODER
	GPIOModeInput         = 0
	GPIOModeOutputGeneral = 1
	GPIOModeOutputAltFunc = 2
	GPIOModeAnalog        = 3

	// GPIOx_OTYPER
	GPIOOutputTypePushPull  = 0
	GPIOOutputTypeOpenDrain = 1

	// GPIOx_OSPEEDR
	GPIOSpeedLow      = 0
	GPIOSpeedMid      = 1
	GPIOSpeedHigh     = 2
	GPIOSpeedVeryHigh = 3

	// GPIOx_PUPDR
	GPIOPUPDRFloating = 0
	GPIOPUPDRPullUp   = 1
	GPIOPUPDRPullDown = 2
)

// Configure this pin with the given configuration.
func (p Pin) configure(config PinConfig) {
	// Configure the GPIO pin.
	p.enableClock()
	port := p.getPort()
	pin := uint8(p) % 16
	pos := pin * 2

	switch config.Mode {

	// GPIO
	case PinInputFloating:
		port.MODER.ReplaceBits(GPIOModeInput, 0x3, pos)
		port.PUPDR.ReplaceBits(GPIOPUPDRFloating, 0x3, pos)
	case PinInputPulldown:
		port.MODER.ReplaceBits(GPIOModeInput, 0x3, pos)
		port.PUPDR.ReplaceBits(GPIOPUPDRPullDown, 0x3, pos)
	case PinInputPullup:
		port.MODER.ReplaceBits(GPIOModeInput, 0x3, pos)
		port.PUPDR.ReplaceBits(GPIOPUPDRPullUp, 0x3, pos)
	case PinOutput:
		port.MODER.ReplaceBits(GPIOModeOutputGeneral, 0x3, pos)
		port.OSPEEDR.ReplaceBits(GPIOSpeedHigh, 0x3, pos)

	// UART
	case PinModeUARTTX:
		port.MODER.ReplaceBits(GPIOModeOutputAltFunc, 0x3, pos)
		port.OSPEEDR.ReplaceBits(GPIOSpeedHigh, 0x3, pos)
		port.PUPDR.ReplaceBits(GPIOPUPDRPullUp, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF7_USART1_2_3)
	case PinModeUARTRX:
		port.MODER.ReplaceBits(GPIOModeOutputAltFunc, 0x3, pos)
		port.PUPDR.ReplaceBits(GPIOPUPDRFloating, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF7_USART1_2_3)

	// I2C
	case PinModeI2CSCL:
		port.MODER.ReplaceBits(GPIOModeOutputAltFunc, 0x3, pos)
		port.OTYPER.ReplaceBits(GPIOOutputTypeOpenDrain, 0x1, pos)
		port.OSPEEDR.ReplaceBits(GPIOSpeedLow, 0x3, pos)
		port.PUPDR.ReplaceBits(GPIOPUPDRFloating, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF4_I2C1_2_3)
	case PinModeI2CSDA:
		port.MODER.ReplaceBits(GPIOModeOutputAltFunc, 0x3, pos)
		port.OTYPER.ReplaceBits(GPIOOutputTypeOpenDrain, 0x1, pos)
		port.OSPEEDR.ReplaceBits(GPIOSpeedLow, 0x3, pos)
		port.PUPDR.ReplaceBits(GPIOPUPDRFloating, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF4_I2C1_2_3)

	// SPI
	case PinModeSPICLK:
		port.MODER.ReplaceBits(GPIOModeOutputAltFunc, 0x3, pos)
		port.OSPEEDR.ReplaceBits(GPIOSpeedLow, 0x3, pos)
		port.PUPDR.ReplaceBits(GPIOPUPDRFloating, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF5_SPI1_SPI2)
	case PinModeSPIMOSI:
		port.MODER.ReplaceBits(GPIOModeOutputAltFunc, 0x3, pos)
		port.OSPEEDR.ReplaceBits(GPIOSpeedLow, 0x3, pos)
		port.PUPDR.ReplaceBits(GPIOPUPDRFloating, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF5_SPI1_SPI2)
	case PinModeSPIMISO:
		port.MODER.ReplaceBits(GPIOModeOutputAltFunc, 0x3, pos)
		port.OSPEEDR.ReplaceBits(GPIOSpeedLow, 0x3, pos)
		port.PUPDR.ReplaceBits(GPIOPUPDRFloating, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF5_SPI1_SPI2)
	}
}

// SetAltFunc maps the given alternative function to the I/O pin
func (p Pin) SetAltFunc(af AltFunc) {
	port := p.getPort()
	pin := uint8(p) % 16
	pos := (pin % 8) * 4
	if pin < 8 {
		port.AFRL.ReplaceBits(uint32(af), 0xf, pos)
	} else {
		port.AFRH.ReplaceBits(uint32(af), 0xf, pos)
	}
}

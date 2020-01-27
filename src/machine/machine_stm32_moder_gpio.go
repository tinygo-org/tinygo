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
	PinModeUartTX PinMode = 4
	PinModeUartRX PinMode = 5

	// for I2C
	PinModeI2CScl PinMode = 6
	PinModeI2CSda PinMode = 7

	// for SPI
	PinModeSpiClk  PinMode = 8
	PinModeSpiMosi PinMode = 9
	PinModeSpiMiso PinMode = 10

	// for analog/ADC
	PinInputAnalog PinMode = 11

	// for PWM

	//GPIOx_MODER
	GPIO_MODE_INPUT          = 0
	GPIO_MODE_GENERAL_OUTPUT = 1
	GPIO_MODE_ALTERNATIVE    = 2
	GPIO_MODE_ANALOG         = 3

	//GPIOx_OTYPER
	GPIO_OUTPUT_MODE_PUSH_PULL  = 0
	GPIO_OUTPUT_MODE_OPEN_DRAIN = 1

	// GPIOx_OSPEEDR
	GPIO_SPEED_LOW     = 0
	GPIO_SPEED_MID     = 1
	GPIO_SPEED_HI      = 2
	GPIO_SPEED_VERY_HI = 3

	// GPIOx_PUPDR
	GPIO_FLOATING  = 0
	GPIO_PULL_UP   = 1
	GPIO_PULL_DOWN = 2
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
		port.MODER.SetMasked(GPIO_MODE_INPUT, 0x3, pos)
		port.PUPDR.SetMasked(GPIO_FLOATING, 0x3, pos)
	case PinInputPulldown:
		port.MODER.SetMasked(GPIO_MODE_INPUT, 0x3, pos)
		port.PUPDR.SetMasked(GPIO_PULL_DOWN, 0x3, pos)
	case PinInputPullup:
		port.MODER.SetMasked(GPIO_MODE_INPUT, 0x3, pos)
		port.PUPDR.SetMasked(GPIO_PULL_UP, 0x3, pos)
	case PinOutput:
		port.MODER.SetMasked(GPIO_MODE_GENERAL_OUTPUT, 0x3, pos)
		port.OSPEEDR.SetMasked(GPIO_SPEED_HI, 0x3, pos)

	// UART
	case PinModeUartTX:
		port.MODER.SetMasked(GPIO_MODE_ALTERNATIVE, 0x3, pos)
		port.OSPEEDR.SetMasked(GPIO_SPEED_HI, 0x3, pos)
		port.PUPDR.SetMasked(GPIO_PULL_UP, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF7_USART1_2_3)
	case PinModeUartRX:
		port.MODER.SetMasked(GPIO_MODE_ALTERNATIVE, 0x3, pos)
		port.PUPDR.SetMasked(GPIO_FLOATING, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF7_USART1_2_3)

	// I2C
	case PinModeI2CScl:
		port.MODER.SetMasked(GPIO_MODE_ALTERNATIVE, 0x3, pos)
		port.OTYPER.SetMasked(GPIO_OUTPUT_MODE_OPEN_DRAIN, 0x1, pos)
		port.OSPEEDR.SetMasked(GPIO_SPEED_LOW, 0x3, pos)
		port.PUPDR.SetMasked(GPIO_FLOATING, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF4_I2C1_2_3)
	case PinModeI2CSda:
		port.MODER.SetMasked(GPIO_MODE_ALTERNATIVE, 0x3, pos)
		port.OTYPER.SetMasked(GPIO_OUTPUT_MODE_OPEN_DRAIN, 0x1, pos)
		port.OSPEEDR.SetMasked(GPIO_SPEED_LOW, 0x3, pos)
		port.PUPDR.SetMasked(GPIO_FLOATING, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF4_I2C1_2_3)

	// SPI
	case PinModeSpiClk:
		port.MODER.SetMasked(GPIO_MODE_ALTERNATIVE, 0x3, pos)
		port.OSPEEDR.SetMasked(GPIO_SPEED_LOW, 0x3, pos)
		port.PUPDR.SetMasked(GPIO_FLOATING, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF5_SPI1_SPI2)
	case PinModeSpiMosi:
		port.MODER.SetMasked(GPIO_MODE_ALTERNATIVE, 0x3, pos)
		port.OSPEEDR.SetMasked(GPIO_SPEED_LOW, 0x3, pos)
		port.PUPDR.SetMasked(GPIO_FLOATING, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF5_SPI1_SPI2)
	case PinModeSpiMiso:
		port.MODER.SetMasked(GPIO_MODE_ALTERNATIVE, 0x3, pos)
		port.OSPEEDR.SetMasked(GPIO_SPEED_LOW, 0x3, pos)
		port.PUPDR.SetMasked(GPIO_FLOATING, 0x3, pos)
		// TODO: determine right alt-func for the peripheral in use
		p.SetAltFunc(AF5_SPI1_SPI2)
	}
}

// SetAltFunc maps the given alternative function to the I/O pin
func (p Pin) SetAltFunc(af AltFunc) {
	port := p.getPort()
	pin := uint8(p) % 16
	pos := pin * 4
	if pin >= 8 {
		port.AFRH.SetMasked(uint32(af), 0xf, pos)
	} else {
		port.AFRL.SetMasked(uint32(af), 0xf, pos)
	}
}

// +build stm32,stm32f103xx

package machine

// Peripheral abstraction layer for the stm32.

import (
	"device/stm32"
	"unsafe"
)

const CPU_FREQUENCY = 72000000

const (
	PinInput       PinMode = 0 // Input mode
	PinOutput10MHz PinMode = 1 // Output mode, max speed 10MHz
	PinOutput2MHz  PinMode = 2 // Output mode, max speed 2MHz
	PinOutput50MHz PinMode = 3 // Output mode, max speed 50MHz
	PinOutput      PinMode = PinOutput2MHz

	PinInputModeAnalog     PinMode = 0  // Input analog mode
	PinInputModeFloating   PinMode = 4  // Input floating mode
	PinInputModePullUpDown PinMode = 8  // Input pull up/down mode
	PinInputModeReserved   PinMode = 12 // Input mode (reserved)

	PinOutputModeGPPushPull   PinMode = 0  // Output mode general purpose push/pull
	PinOutputModeGPOpenDrain  PinMode = 4  // Output mode general purpose open drain
	PinOutputModeAltPushPull  PinMode = 8  // Output mode alt. purpose push/pull
	PinOutputModeAltOpenDrain PinMode = 12 // Output mode alt. purpose open drain
)

// Configure this pin with the given I/O settings.
// stm32f1xx uses different technique for setting the GPIO pins than the stm32f407
func (p Pin) configure(config PinConfig) {
	// Configure the GPIO pin.
	port := p.getPort()
	pin := uint8(p) % 16
	pos := uint8(p) % 8 * 4
	if pin < 8 {
		port.CRL.SetMasked(uint32(config.Mode), 0xf, pos)
	} else {
		port.CRH.SetMasked(uint32(config.Mode), 0xf, pos)
	}
}

func (p Pin) getPort() *stm32.GPIO_Type {
	switch p / 16 {
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

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPAEN)
	case 1:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPBEN)
	case 2:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPCEN)
	case 3:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPDEN)
	case 4:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPEEN)
	case 5:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPFEN)
	case 6:
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_IOPGEN)
	default:
		panic("machine: unknown port")
	}
}

// Enable peripheral clock. Expand to include all the desired peripherals
func enableAltFuncClock(bus unsafe.Pointer) {
	if bus == unsafe.Pointer(stm32.USART1) {
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	} else if bus == unsafe.Pointer(stm32.USART2) {
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_USART2EN)
	} else if bus == unsafe.Pointer(stm32.I2C1) {
		stm32.RCC.APB1ENR.SetBits(stm32.RCC_APB1ENR_I2C1EN)
	} else if bus == unsafe.Pointer(stm32.SPI1) {
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	}
}

// Configure the TX and RX pins
func (uart UART) configurePins(config UARTConfig) {

	// pins
	if config.TX == UART_ALT_TX_PIN {
		// use alternate TX/RX pins via AFIO mapping
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_AFIOEN)
		if uart.Bus == stm32.USART1 {
			stm32.AFIO.MAPR.SetBits(stm32.AFIO_MAPR_USART1_REMAP)
		} else if uart.Bus == stm32.USART2 {
			stm32.AFIO.MAPR.SetBits(stm32.AFIO_MAPR_USART2_REMAP)
		}
	}

	config.TX.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltPushPull})
	config.RX.Configure(PinConfig{Mode: PinInputModeFloating})
}

// Determine the divisor for USARTs to get the given baudrate
func (uart UART) getBaudRateDivisor(br uint32) uint32 {

	// Note: PCLK2 (from APB2) used for USART1 and PCLK1 for USART2, 3, 4, 5
	var divider uint32
	if uart.Bus == stm32.USART1 {
		// first divide by PCLK2 prescaler (div 1) and then desired baudrate
		divider = CPU_FREQUENCY / br
	} else {
		// first divide by PCLK1 prescaler (div 2) and then desired baudrate
		divider = CPU_FREQUENCY / 2 / br
	}
	return divider
}

// Configure the SDA and SCL pins for I2C
func (i2c I2C) configurePins(config I2CConfig) {
	if i2c.Bus == stm32.I2C1 {
		switch config.SDA {
		case PB9:
			config.SCL = PB8
			// use alternate I2C1 pins PB8/PB9 via AFIO mapping
			stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_AFIOEN)
			stm32.AFIO.MAPR.SetBits(stm32.AFIO_MAPR_I2C1_REMAP)
		default:
			// use default I2C1 pins PB6/PB7
			config.SDA = SDA_PIN
			config.SCL = SCL_PIN
		}
	}

	config.SDA.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltOpenDrain})
	config.SCL.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltOpenDrain})
}

// Set baud rate for SPI
func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var conf uint32

	// set frequency dependent on PCLK2 prescaler (div 1)
	switch config.Frequency {
	case 125000:
		// Note: impossible to achieve lower frequency with current PCLK2!
		conf |= stm32.SPI_BaudRatePrescaler_256
	case 250000:
		conf |= stm32.SPI_BaudRatePrescaler_256
	case 500000:
		conf |= stm32.SPI_BaudRatePrescaler_128
	case 1000000:
		conf |= stm32.SPI_BaudRatePrescaler_64
	case 2000000:
		conf |= stm32.SPI_BaudRatePrescaler_32
	case 4000000:
		conf |= stm32.SPI_BaudRatePrescaler_16
	case 8000000:
		conf |= stm32.SPI_BaudRatePrescaler_8
	default:
		conf |= stm32.SPI_BaudRatePrescaler_256
	}
	return conf
}

// Configure SPI pins for input output and clock
func (spi SPI) configurePins(config SPIConfig) {
	if config.SCK == 0 {
		config.SCK = SPI0_SCK_PIN
	}
	if config.MOSI == 0 {
		config.MOSI = SPI0_MOSI_PIN
	}
	if config.MISO == 0 {
		config.MISO = SPI0_MISO_PIN
	}

	config.SCK.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltPushPull})
	config.MOSI.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltPushPull})
	config.MISO.Configure(PinConfig{Mode: PinInputModeFloating})
}

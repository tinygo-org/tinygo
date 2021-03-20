// +build stm32,stm32f103

package machine

// Peripheral abstraction layer for the stm32.

import (
	"device/stm32"
	"unsafe"
)

func CPUFrequency() uint32 {
	return 72000000
}

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
func (p Pin) Configure(config PinConfig) {
	// Configure the GPIO pin.
	p.enableClock()
	port := p.getPort()
	pin := uint8(p) % 16
	pos := (pin % 8) * 4
	if pin < 8 {
		port.CRL.ReplaceBits(uint32(config.Mode), 0xf, pos)
	} else {
		port.CRH.ReplaceBits(uint32(config.Mode), 0xf, pos)
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

//---------- UART related code

// Configure the TX and RX pins
func (uart *UART) configurePins(config UARTConfig) {

	// pins
	switch config.TX {
	case UART_ALT_TX_PIN:
		// use alternate TX/RX pins via AFIO mapping
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_AFIOEN)
		if uart.Bus == stm32.USART1 {
			stm32.AFIO.MAPR.SetBits(stm32.AFIO_MAPR_USART1_REMAP)
		} else if uart.Bus == stm32.USART2 {
			stm32.AFIO.MAPR.SetBits(stm32.AFIO_MAPR_USART2_REMAP)
		}
	default:
		// use standard TX/RX pins PA9 and PA10
	}
	config.TX.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltPushPull})
	config.RX.Configure(PinConfig{Mode: PinInputModeFloating})
}

// Determine the divisor for USARTs to get the given baudrate
func (uart *UART) getBaudRateDivisor(br uint32) uint32 {

	// Note: PCLK2 (from APB2) used for USART1 and PCLK1 for USART2, 3, 4, 5
	var divider uint32
	if uart.Bus == stm32.USART1 {
		// first divide by PCLK2 prescaler (div 1) and then desired baudrate
		divider = CPUFrequency() / br
	} else {
		// first divide by PCLK1 prescaler (div 2) and then desired baudrate
		divider = CPUFrequency() / 2 / br
	}
	return divider
}

// Register names vary by ST processor, these are for STM F103xx
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.DR
	uart.txReg = &uart.Bus.DR
	uart.statusReg = &uart.Bus.SR
	uart.txEmptyFlag = stm32.USART_SR_TXE
}

//---------- SPI related types and code

type SPI struct {
	Bus *stm32.SPI_Type
}

// There are 3 SPI interfaces on the STM32F103xx.
// Since the first interface is named SPI1, both SPI0 and SPI1 refer to SPI1.
// TODO: implement SPI2 and SPI3.
var (
	SPI1 = SPI{Bus: stm32.SPI1}
	SPI0 = SPI1
)

// Set baud rate for SPI
func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var conf uint32

	// set frequency dependent on PCLK2 prescaler (div 1)
	switch config.Frequency {
	case 125000:
		// Note: impossible to achieve lower frequency with current PCLK2!
		conf |= stm32.SPI_CR1_BR_Div256
	case 250000:
		conf |= stm32.SPI_CR1_BR_Div256
	case 500000:
		conf |= stm32.SPI_CR1_BR_Div128
	case 1000000:
		conf |= stm32.SPI_CR1_BR_Div64
	case 2000000:
		conf |= stm32.SPI_CR1_BR_Div32
	case 4000000:
		conf |= stm32.SPI_CR1_BR_Div16
	case 8000000:
		conf |= stm32.SPI_CR1_BR_Div8
	default:
		conf |= stm32.SPI_CR1_BR_Div256
	}
	return conf
}

// Configure SPI pins for input output and clock
func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltPushPull})
	config.SDO.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltPushPull})
	config.SDI.Configure(PinConfig{Mode: PinInputModeFloating})
}

//---------- I2C related types and code

// There are 2 I2C interfaces on the STM32F103xx.
// Since the first interface is named I2C1, both I2C0 and I2C1 refer to I2C1.
// TODO: implement I2C2.
var (
	I2C1 = I2C{Bus: stm32.I2C1}
	I2C0 = I2C1
)

type I2C struct {
	Bus *stm32.I2C_Type
}

func (i2c I2C) configurePins(config I2CConfig) {
	if config.SDA == PB9 {
		// use alternate I2C1 pins PB8/PB9 via AFIO mapping
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_AFIOEN)
		stm32.AFIO.MAPR.SetBits(stm32.AFIO_MAPR_I2C1_REMAP)
	}

	config.SDA.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltOpenDrain})
	config.SCL.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltOpenDrain})
}

func (i2c I2C) getFreqRange(config I2CConfig) uint32 {
	// pclk1 clock speed is main frequency divided by PCLK1 prescaler (div 2)
	pclk1 := CPUFrequency() / 2

	// set freqency range to PCLK1 clock speed in MHz
	// aka setting the value 36 means to use 36 MHz clock
	return pclk1 / 1000000
}

func (i2c I2C) getRiseTime(config I2CConfig) uint32 {
	// These bits must be programmed with the maximum SCL rise time given in the
	// I2C bus specification, incremented by 1.
	// For instance: in Sm mode, the maximum allowed SCL rise time is 1000 ns.
	// If, in the I2C_CR2 register, the value of FREQ[5:0] bits is equal to 0x08
	// and PCLK1 = 125 ns, therefore the TRISE[5:0] bits must be programmed with
	// 09h (1000 ns / 125 ns = 8 + 1)
	freqRange := i2c.getFreqRange(config)
	if config.Frequency > 100000 {
		// fast mode (Fm) adjustment
		freqRange *= 300
		freqRange /= 1000
	}
	return (freqRange + 1) << stm32.I2C_TRISE_TRISE_Pos
}

func (i2c I2C) getSpeed(config I2CConfig) uint32 {
	ccr := func(pclk uint32, freq uint32, coeff uint32) uint32 {
		return (((pclk - 1) / (freq * coeff)) + 1) & stm32.I2C_CCR_CCR_Msk
	}
	sm := func(pclk uint32, freq uint32) uint32 { // standard mode (Sm)
		if s := ccr(pclk, freq, 2); s < 4 {
			return 4
		} else {
			return s
		}
	}
	fm := func(pclk uint32, freq uint32, duty uint8) uint32 { // fast mode (Fm)
		if duty == DutyCycle2 {
			return ccr(pclk, freq, 3)
		} else {
			return ccr(pclk, freq, 25) | stm32.I2C_CCR_DUTY
		}
	}
	clock := CPUFrequency() / 2
	if config.Frequency <= 100000 {
		return sm(clock, config.Frequency)
	} else {
		s := fm(clock, config.Frequency, config.DutyCycle)
		if (s & stm32.I2C_CCR_CCR_Msk) == 0 {
			return 1
		} else {
			return s | stm32.I2C_CCR_F_S
		}
	}
}

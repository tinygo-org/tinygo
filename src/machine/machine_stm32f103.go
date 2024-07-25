//go:build stm32 && stm32f103

package machine

// Peripheral abstraction layer for the stm32.

import (
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

func CPUFrequency() uint32 {
	return 72000000
}

var deviceIDAddr = []uintptr{0x1FFFF7E8, 0x1FFFF7EC, 0x1FFFF7F0}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 72e6 // 72MHz
const APB2_TIM_FREQ = 72e6 // 72MHz

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

	// Pull-up vs Pull down is not part of the CNF0 / CNF1 bits, but is
	// controlled by PxODR.  Encoded using the 'spare' bit 5.
	PinInputPulldown PinMode = PinInputModePullUpDown
	PinInputPullup   PinMode = PinInputModePullUpDown | 0x10
)

// Pin constants for all stm32f103 package sizes
const (
	PA0  = portA + 0
	PA1  = portA + 1
	PA2  = portA + 2
	PA3  = portA + 3
	PA4  = portA + 4
	PA5  = portA + 5
	PA6  = portA + 6
	PA7  = portA + 7
	PA8  = portA + 8
	PA9  = portA + 9
	PA10 = portA + 10
	PA11 = portA + 11
	PA12 = portA + 12
	PA13 = portA + 13
	PA14 = portA + 14
	PA15 = portA + 15

	PB0  = portB + 0
	PB1  = portB + 1
	PB2  = portB + 2
	PB3  = portB + 3
	PB4  = portB + 4
	PB5  = portB + 5
	PB6  = portB + 6
	PB7  = portB + 7
	PB8  = portB + 8
	PB9  = portB + 9
	PB10 = portB + 10
	PB11 = portB + 11
	PB12 = portB + 12
	PB13 = portB + 13
	PB14 = portB + 14
	PB15 = portB + 15

	PC0  = portC + 0
	PC1  = portC + 1
	PC2  = portC + 2
	PC3  = portC + 3
	PC4  = portC + 4
	PC5  = portC + 5
	PC6  = portC + 6
	PC7  = portC + 7
	PC8  = portC + 8
	PC9  = portC + 9
	PC10 = portC + 10
	PC11 = portC + 11
	PC12 = portC + 12
	PC13 = portC + 13
	PC14 = portC + 14
	PC15 = portC + 15

	PD0  = portD + 0
	PD1  = portD + 1
	PD2  = portD + 2
	PD3  = portD + 3
	PD4  = portD + 4
	PD5  = portD + 5
	PD6  = portD + 6
	PD7  = portD + 7
	PD8  = portD + 8
	PD9  = portD + 9
	PD10 = portD + 10
	PD11 = portD + 11
	PD12 = portD + 12
	PD13 = portD + 13
	PD14 = portD + 14
	PD15 = portD + 15

	PE0  = portE + 0
	PE1  = portE + 1
	PE2  = portE + 2
	PE3  = portE + 3
	PE4  = portE + 4
	PE5  = portE + 5
	PE6  = portE + 6
	PE7  = portE + 7
	PE8  = portE + 8
	PE9  = portE + 9
	PE10 = portE + 10
	PE11 = portE + 11
	PE12 = portE + 12
	PE13 = portE + 13
	PE14 = portE + 14
	PE15 = portE + 15

	PF0  = portF + 0
	PF1  = portF + 1
	PF2  = portF + 2
	PF3  = portF + 3
	PF4  = portF + 4
	PF5  = portF + 5
	PF6  = portF + 6
	PF7  = portF + 7
	PF8  = portF + 8
	PF9  = portF + 9
	PF10 = portF + 10
	PF11 = portF + 11
	PF12 = portF + 12
	PF13 = portF + 13
	PF14 = portF + 14
	PF15 = portF + 15
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

	// If configured for input pull-up or pull-down, set ODR
	// for desired pull-up or pull-down.
	if (config.Mode & 0xf) == PinInputModePullUpDown {
		var pullup uint32
		if config.Mode == PinInputPullup {
			pullup = 1
		}
		port.ODR.ReplaceBits(pullup, 0x1, pin)
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

func (p Pin) registerInterrupt() interrupt.Interrupt {
	pin := uint8(p) % 16

	switch pin {
	case 0:
		return interrupt.New(stm32.IRQ_EXTI0, func(interrupt.Interrupt) { handlePinInterrupt(0) })
	case 1:
		return interrupt.New(stm32.IRQ_EXTI1, func(interrupt.Interrupt) { handlePinInterrupt(1) })
	case 2:
		return interrupt.New(stm32.IRQ_EXTI2, func(interrupt.Interrupt) { handlePinInterrupt(2) })
	case 3:
		return interrupt.New(stm32.IRQ_EXTI3, func(interrupt.Interrupt) { handlePinInterrupt(3) })
	case 4:
		return interrupt.New(stm32.IRQ_EXTI4, func(interrupt.Interrupt) { handlePinInterrupt(4) })
	case 5:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(5) })
	case 6:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(6) })
	case 7:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(7) })
	case 8:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(8) })
	case 9:
		return interrupt.New(stm32.IRQ_EXTI9_5, func(interrupt.Interrupt) { handlePinInterrupt(9) })
	case 10:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(10) })
	case 11:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(11) })
	case 12:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(12) })
	case 13:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(13) })
	case 14:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(14) })
	case 15:
		return interrupt.New(stm32.IRQ_EXTI15_10, func(interrupt.Interrupt) { handlePinInterrupt(15) })
	}

	return interrupt.Interrupt{}
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

func (spi SPI) config8Bits() {
	// no-op on this series
}

// Set baud rate for SPI
func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var conf uint32

	// set frequency dependent on PCLK2 prescaler (div 1)
	switch {
	case config.Frequency < 125000:
		// Note: impossible to achieve lower frequency with current PCLK2!
		conf |= stm32.SPI_CR1_BR_Div256
	case config.Frequency < 250000:
		conf |= stm32.SPI_CR1_BR_Div256
	case config.Frequency < 500000:
		conf |= stm32.SPI_CR1_BR_Div128
	case config.Frequency < 1000000:
		conf |= stm32.SPI_CR1_BR_Div64
	case config.Frequency < 2000000:
		conf |= stm32.SPI_CR1_BR_Div32
	case config.Frequency < 4000000:
		conf |= stm32.SPI_CR1_BR_Div16
	default:
		// When its bigger than Div16, just round to the maximum frequency.
		conf |= stm32.SPI_CR1_BR_Div8
	}
	return conf << stm32.SPI_CR1_BR_Pos
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

type I2C struct {
	Bus *stm32.I2C_Type
}

var (
	I2C1 = &I2C{Bus: stm32.I2C1}
	I2C0 = I2C1
)

func (i2c *I2C) configurePins(config I2CConfig) {
	if config.SDA == PB9 {
		// use alternate I2C1 pins PB8/PB9 via AFIO mapping
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_AFIOEN)
		stm32.AFIO.MAPR.SetBits(stm32.AFIO_MAPR_I2C1_REMAP)
	}

	config.SDA.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltOpenDrain})
	config.SCL.Configure(PinConfig{Mode: PinOutput50MHz + PinOutputModeAltOpenDrain})
}

func (i2c *I2C) getFreqRange(config I2CConfig) uint32 {
	// pclk1 clock speed is main frequency divided by PCLK1 prescaler (div 2)
	pclk1 := CPUFrequency() / 2

	// set frequency range to PCLK1 clock speed in MHz
	// aka setting the value 36 means to use 36 MHz clock
	return pclk1 / 1000000
}

func (i2c *I2C) getRiseTime(config I2CConfig) uint32 {
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

func (i2c *I2C) getSpeed(config I2CConfig) uint32 {
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

//---------- Timer related code

// For Pin Mappings see RM0008, pg 179
// https://www.st.com/resource/en/reference_manual/cd00171190-stm32f101xx-stm32f102xx-stm32f103xx-stm32f105xx-and-stm32f107xx-advanced-arm-based-32-bit-mcus-stmicroelectronics.pdf
//
// Note: for STM32F1 series the pin mapping is done 'per timer' not per channel,
// not all channels on a timer have the same degrees of flexibility, and some
// combinations are only available on some packages - so care is needed at app
// level to ensure valid combinations of pins are used.
//

var (
	TIM1 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM1EN,
		Device:         stm32.TIM1,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PE9, 0b11}, {PA8, 0b00}}},
			TimerChannel{Pins: []PinFunction{{PE11, 0b11}, {PA9, 0b00}}},
			TimerChannel{Pins: []PinFunction{{PE13, 0b11}, {PA10, 0b00}}},
			TimerChannel{Pins: []PinFunction{{PE14, 0b11}, {PA11, 0b00}}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM2 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM2EN,
		Device:         stm32.TIM2,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA0, 0b00}, {PA15, 0b01}}},
			TimerChannel{Pins: []PinFunction{{PA1, 0b00}, {PB3, 0b01}}},
			TimerChannel{Pins: []PinFunction{{PA2, 0b00}, {PB10, 0b10}}},
			TimerChannel{Pins: []PinFunction{{PA3, 0b00}, {PB11, 0b10}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM3 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM3EN,
		Device:         stm32.TIM3,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA6, 0b00}, {PC6, 0b11}, {PB4, 0b10}}},
			TimerChannel{Pins: []PinFunction{{PA7, 0b00}, {PC7, 0b11}, {PB5, 0b10}}},
			TimerChannel{Pins: []PinFunction{{PB0, 0b00}, {PC8, 0b11}}},
			TimerChannel{Pins: []PinFunction{{PB1, 0b00}, {PC9, 0b11}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM4 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM4EN,
		Device:         stm32.TIM4,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PD12, 0b1}, {PB6, 0}}},
			TimerChannel{Pins: []PinFunction{{PD13, 0b1}, {PB7, 0}}},
			TimerChannel{Pins: []PinFunction{{PD14, 0b1}, {PB8, 0}}},
			TimerChannel{Pins: []PinFunction{{PD15, 0b1}, {PB9, 0}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM5 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM5EN,
		Device:         stm32.TIM5,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{{PA3, 0b0}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM6 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM6EN,
		Device:         stm32.TIM6,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM7 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM7EN,
		Device:         stm32.TIM7,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM8 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM8EN,
		Device:         stm32.TIM8,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM9 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM9EN,
		Device:         stm32.TIM9,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA2, 0b0}, {PE5, 0b1}}},
			TimerChannel{Pins: []PinFunction{{PA3, 0b0}, {PE6, 0b1}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM10 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM10EN,
		Device:         stm32.TIM10,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PB8, 0b0}, {PF6, 0b1}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM11 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM11EN,
		Device:         stm32.TIM11,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PB9, 0b0}, {PF7, 0b1}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM12 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM12EN,
		Device:         stm32.TIM12,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM13 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM13EN,
		Device:         stm32.TIM13,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA6, 0b0}, {PF8, 0b1}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM14 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR,
		EnableFlag:     stm32.RCC_APB1ENR_TIM14EN,
		Device:         stm32.TIM14,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA7, 0b0}, {PF9, 0b1}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}
)

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(stm32.IRQ_TIM1_UP, TIM1.handleUPInterrupt)
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleUPInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleUPInterrupt)
	case &TIM4:
		return interrupt.New(stm32.IRQ_TIM4, TIM4.handleUPInterrupt)
	case &TIM5:
		return interrupt.New(stm32.IRQ_TIM5, TIM5.handleUPInterrupt)
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6, TIM6.handleUPInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7, TIM7.handleUPInterrupt)
	case &TIM8:
		return interrupt.New(stm32.IRQ_TIM8_UP, TIM8.handleUPInterrupt)
	}

	return interrupt.Interrupt{}
}

func (t *TIM) registerOCInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(stm32.IRQ_TIM1_CC, TIM1.handleOCInterrupt)
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleOCInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3, TIM3.handleOCInterrupt)
	case &TIM4:
		return interrupt.New(stm32.IRQ_TIM4, TIM4.handleOCInterrupt)
	case &TIM5:
		return interrupt.New(stm32.IRQ_TIM5, TIM5.handleOCInterrupt)
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6, TIM6.handleOCInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7, TIM7.handleOCInterrupt)
	case &TIM8:
		return interrupt.New(stm32.IRQ_TIM8_CC, TIM8.handleOCInterrupt)
	}

	return interrupt.Interrupt{}
}

func (t *TIM) configurePin(channel uint8, pf PinFunction) {
	remap := uint32(pf.AltFunc)

	switch t {
	case &TIM1:
		stm32.AFIO.MAPR.ReplaceBits(remap<<stm32.AFIO_MAPR_TIM1_REMAP_Pos, stm32.AFIO_MAPR_TIM1_REMAP_Msk, 0)
	case &TIM2:
		stm32.AFIO.MAPR.ReplaceBits(remap<<stm32.AFIO_MAPR_TIM2_REMAP_Pos, stm32.AFIO_MAPR_TIM2_REMAP_Msk, 0)
	case &TIM3:
		stm32.AFIO.MAPR.ReplaceBits(remap<<stm32.AFIO_MAPR_TIM3_REMAP_Pos, stm32.AFIO_MAPR_TIM3_REMAP_Msk, 0)
	case &TIM4:
		stm32.AFIO.MAPR.ReplaceBits(remap<<stm32.AFIO_MAPR_TIM4_REMAP_Pos, stm32.AFIO_MAPR_TIM4_REMAP_Msk, 0)
	case &TIM5:
		stm32.AFIO.MAPR.ReplaceBits(remap<<stm32.AFIO_MAPR_TIM5CH4_IREMAP_Pos, stm32.AFIO_MAPR_TIM5CH4_IREMAP_Msk, 0)
	case &TIM9:
		stm32.AFIO.MAPR.ReplaceBits(remap<<stm32.AFIO_MAPR2_TIM9_REMAP_Pos, stm32.AFIO_MAPR2_TIM9_REMAP_Msk, 0)
	case &TIM10:
		stm32.AFIO.MAPR.ReplaceBits(remap<<stm32.AFIO_MAPR2_TIM10_REMAP_Pos, stm32.AFIO_MAPR2_TIM10_REMAP_Msk, 0)
	case &TIM11:
		stm32.AFIO.MAPR.ReplaceBits(remap<<stm32.AFIO_MAPR2_TIM11_REMAP_Pos, stm32.AFIO_MAPR2_TIM11_REMAP_Msk, 0)
	case &TIM13:
		stm32.AFIO.MAPR.ReplaceBits(remap<<stm32.AFIO_MAPR2_TIM13_REMAP_Pos, stm32.AFIO_MAPR2_TIM13_REMAP_Msk, 0)
	case &TIM14:
		stm32.AFIO.MAPR.ReplaceBits(remap<<stm32.AFIO_MAPR2_TIM14_REMAP_Pos, stm32.AFIO_MAPR2_TIM14_REMAP_Msk, 0)
	}

	pf.Pin.Configure(PinConfig{Mode: PinOutput + PinOutputModeAltPushPull})
}

func (t *TIM) enableMainOutput() {
	t.Device.BDTR.SetBits(stm32.TIM_BDTR_MOE)
}

type arrtype = uint32
type arrRegType = volatile.Register32

const (
	ARR_MAX = 0x10000
	PSC_MAX = 0x10000
)

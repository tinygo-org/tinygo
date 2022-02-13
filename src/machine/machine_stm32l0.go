//go:build stm32l0
// +build stm32l0

package machine

// Peripheral abstraction layer for the stm32l0

import (
	"runtime/interrupt"
	"tinygo.org/x/device/stm32"
)

func CPUFrequency() uint32 {
	return 32000000
}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 32e6 // 32MHz
const APB2_TIM_FREQ = 32e6 // 32MHz

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

	PH0 = portH + 0
	PH1 = portH + 1
)

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
	case 7:
		return stm32.GPIOH
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPAEN)
	case 1:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPBEN)
	case 2:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPCEN)
	case 3:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPDEN)
	case 4:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPEEN)
	case 7:
		stm32.RCC.IOPENR.SetBits(stm32.RCC_IOPENR_IOPHEN)
	default:
		panic("machine: unknown port")
	}
}

func (p Pin) registerInterrupt() interrupt.Interrupt {
	pin := uint8(p) % 16

	switch pin {
	case 0:
		return interrupt.New(stm32.IRQ_EXTI0_1, func(interrupt.Interrupt) { handlePinInterrupt(0) })
	case 1:
		return interrupt.New(stm32.IRQ_EXTI0_1, func(interrupt.Interrupt) { handlePinInterrupt(1) })
	case 2:
		return interrupt.New(stm32.IRQ_EXTI2_3, func(interrupt.Interrupt) { handlePinInterrupt(2) })
	case 3:
		return interrupt.New(stm32.IRQ_EXTI2_3, func(interrupt.Interrupt) { handlePinInterrupt(3) })
	case 4:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(4) })
	case 5:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(5) })
	case 6:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(6) })
	case 7:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(7) })
	case 8:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(8) })
	case 9:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(9) })
	case 10:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(10) })
	case 11:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(11) })
	case 12:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(12) })
	case 13:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(13) })
	case 14:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(14) })
	case 15:
		return interrupt.New(stm32.IRQ_EXTI4_15, func(interrupt.Interrupt) { handlePinInterrupt(15) })
	}

	return interrupt.Interrupt{}
}

//---------- UART related types and code

// Configure the UART.
func (uart *UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.TxAltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.RxAltFuncSelector)
}

// UART baudrate calc based on the bus and clockspeed
func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	var clock, rate uint32
	switch uart.Bus {
	case stm32.LPUART1:
		clock = CPUFrequency() / 2 // APB1 Frequency
		rate = uint32((256 * clock) / baudRate)
	case stm32.USART1:
		clock = CPUFrequency() / 2 // APB2 Frequency
		rate = uint32(clock / baudRate)
	case stm32.USART2:
		clock = CPUFrequency() / 2 // APB1 Frequency
		rate = uint32(clock / baudRate)
	}

	return rate
}

// Register names vary by ST processor, these are for STM L0 family
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.RDR
	uart.txReg = &uart.Bus.TDR
	uart.statusReg = &uart.Bus.ISR
	uart.txEmptyFlag = stm32.USART_ISR_TXE
}

//---------- SPI related types and code

// SPI on the STM32Fxxx using MODER / alternate function pins
type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
}

func (spi SPI) config8Bits() {
	// no-op on this series
}

// Set baud rate for SPI
func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var conf uint32

	localFrequency := config.Frequency

	// Default
	if config.Frequency == 0 {
		config.Frequency = 4e6
	}

	if spi.Bus != stm32.SPI1 {
		// Assume it's SPI2 or SPI3 on APB1 at 1/2 the clock frequency of APB2, so
		//  we want to pretend to request 2x the baudrate asked for
		localFrequency = localFrequency * 2
	}

	// set frequency dependent on PCLK prescaler. Since these are rather weird
	// speeds due to the CPU freqency, pick a range up to that frquency for
	// clients to use more human-understandable numbers, e.g. nearest 100KHz

	// These are based on APB2 clock frquency (84MHz on the discovery board)
	// TODO: also include the MCU/APB clock setting in the equation
	switch {
	case localFrequency < 328125:
		conf = stm32.SPI_CR1_BR_Div256
	case localFrequency < 656250:
		conf = stm32.SPI_CR1_BR_Div128
	case localFrequency < 1312500:
		conf = stm32.SPI_CR1_BR_Div64
	case localFrequency < 2625000:
		conf = stm32.SPI_CR1_BR_Div32
	case localFrequency < 5250000:
		conf = stm32.SPI_CR1_BR_Div16
	case localFrequency < 10500000:
		conf = stm32.SPI_CR1_BR_Div8
		// NOTE: many SPI components won't operate reliably (or at all) above 10MHz
		// Check the datasheet of the part
	case localFrequency < 21000000:
		conf = stm32.SPI_CR1_BR_Div4
	case localFrequency < 42000000:
		conf = stm32.SPI_CR1_BR_Div2
	default:
		// None of the specific baudrates were selected; choose the lowest speed
		conf = stm32.SPI_CR1_BR_Div256
	}

	return conf << stm32.SPI_CR1_BR_Pos
}

// Configure SPI pins for input output and clock
func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}

//---------- I2C related types and code

// Gets the value for TIMINGR register
func (i2c I2C) getFreqRange() uint32 {
	// This is a 'magic' value calculated by STM32CubeMX
	// for 80MHz PCLK1.
	// TODO: Do calculations based on PCLK1
	return 0x00303D5B
}

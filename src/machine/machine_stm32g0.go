//go:build stm32g0

package machine

// Peripheral abstraction layer for the stm32g0

import (
	"device/stm32"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const (
	// CPU frequency for STM32G0 (64MHz via PLL: HSI16 / 1 * 8 / 2)
	cpuFreq = 64000000
)

func CPUFrequency() uint32 {
	return cpuFreq
}

var deviceIDAddr = []uintptr{0x1FFF7590, 0x1FFF7594, 0x1FFF7598}

// Internal use: configured speed of the APB1 and APB2 timers, this should be kept
// in sync with any changes to runtime package which configures the oscillators
// and clock frequencies
const APB1_TIM_FREQ = 64e6 // 64MHz (PLL: HSI16 / 1 * 8 / 2)
const APB2_TIM_FREQ = 64e6 // 64MHz (PLL: HSI16 / 1 * 8 / 2)

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
	default:
		panic("machine: unknown port")
	}
}

// enableClock enables the clock for this desired GPIO port.
func (p Pin) enableClock() {
	switch p / 16 {
	case 0:
		stm32.RCC.SetIOPENR_GPIOAEN(1)
	case 1:
		stm32.RCC.SetIOPENR_GPIOBEN(1)
	case 2:
		stm32.RCC.SetIOPENR_GPIOCEN(1)
	case 3:
		stm32.RCC.SetIOPENR_GPIODEN(1)
	case 4:
		stm32.RCC.SetIOPENR_GPIOEEN(1)
	case 5:
		stm32.RCC.SetIOPENR_GPIOFEN(1)
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
	return CPUFrequency() / baudRate
}

// Register names vary by ST processor, these are for STM G0 family
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.RDR
	uart.txReg = &uart.Bus.TDR
	uart.statusReg = &uart.Bus.ISR
	uart.txEmptyFlag = stm32.USART_ISR_TXE
	uart.errClearReg = &uart.Bus.ICR
}

//---------- SPI related types and code

// SPI on the STM32G0 using MODER / alternate function pins
type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
}

func (spi *SPI) config8Bits() {
	// Set rx threshold to 8-bits, so RXNE flag is set for 1 byte
	spi.Bus.SetCR2_FRXTH(1)
}

// Set baud rate for SPI
func (spi *SPI) getBaudRate(config SPIConfig) uint32 {
	var conf uint32

	localFrequency := config.Frequency

	// Default
	if localFrequency == 0 {
		localFrequency = 4e6
	}

	// set frequency dependent on PCLK prescaler
	switch {
	case localFrequency < 250000:
		conf = stm32.SPI_CR1_BR_Div256
	case localFrequency < 500000:
		conf = stm32.SPI_CR1_BR_Div128
	case localFrequency < 1000000:
		conf = stm32.SPI_CR1_BR_Div64
	case localFrequency < 2000000:
		conf = stm32.SPI_CR1_BR_Div32
	case localFrequency < 4000000:
		conf = stm32.SPI_CR1_BR_Div16
	case localFrequency < 8000000:
		conf = stm32.SPI_CR1_BR_Div8
	case localFrequency < 16000000:
		conf = stm32.SPI_CR1_BR_Div4
	case localFrequency < 32000000:
		conf = stm32.SPI_CR1_BR_Div2
	default:
		// None of the specific baudrates were selected; choose the lowest speed
		conf = stm32.SPI_CR1_BR_Div256
	}

	return conf << stm32.SPI_CR1_BR_Pos
}

// Configure SPI pins for input output and clock
func (spi *SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}

//---------- I2C related types and code

// Gets the value for TIMINGR register
func (i2c *I2C) getFreqRange(br uint32) uint32 {
	// These are 'magic' values calculated by STM32CubeMX
	// for 64MHz PCLK1 (PLL: HSI16 / 1 * 8 / 2).
	// TODO: Do calculations based on PCLK1
	switch br {
	case 10 * KHz:
		return 0xF010F3FE // 64MHz, 10kHz I2C
	case 100 * KHz:
		return 0x30A0A7FB // 64MHz, 100kHz I2C (Standard mode)
	case 400 * KHz:
		return 0x10802D9B // 64MHz, 400kHz I2C (Fast mode)
	case 500 * KHz:
		return 0x00802172 // 64MHz, 500kHz I2C
	default:
		return 0
	}
}

// Enable peripheral clock
func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	case unsafe.Pointer(stm32.PWR): // Power interface clock enable
		stm32.RCC.SetAPBENR1_PWREN(1)
	case unsafe.Pointer(stm32.I2C1): // I2C1 clock enable
		stm32.RCC.SetAPBENR1_I2C1EN(1)
	case unsafe.Pointer(stm32.I2C2): // I2C2 clock enable
		stm32.RCC.SetAPBENR1_I2C2EN(1)
	case unsafe.Pointer(stm32.USART2): // USART2 clock enable
		stm32.RCC.SetAPBENR1_USART2EN(1)
	case unsafe.Pointer(stm32.USART3): // USART3 clock enable
		stm32.RCC.SetAPBENR1_USART3EN(1)
	case unsafe.Pointer(stm32.USART4): // USART4 clock enable
		stm32.RCC.SetAPBENR1_USART4EN(1)
	case unsafe.Pointer(stm32.SPI2): // SPI2 clock enable
		stm32.RCC.SetAPBENR1_SPI2EN(1)
	case unsafe.Pointer(stm32.WWDG): // Window watchdog clock enable
		stm32.RCC.SetAPBENR1_WWDGEN(1)
	case unsafe.Pointer(stm32.TIM2): // TIM2 clock enable
		stm32.RCC.SetAPBENR1_TIM2EN(1)
	case unsafe.Pointer(stm32.TIM3): // TIM3 clock enable
		stm32.RCC.SetAPBENR1_TIM3EN(1)
	case unsafe.Pointer(stm32.TIM6): // TIM6 clock enable
		stm32.RCC.SetAPBENR1_TIM6EN(1)
	case unsafe.Pointer(stm32.TIM7): // TIM7 clock enable
		stm32.RCC.SetAPBENR1_TIM7EN(1)
	case unsafe.Pointer(stm32.LPUART1): // LPUART1 clock enable
		stm32.RCC.SetAPBENR1_LPUART1EN(1)
	case unsafe.Pointer(stm32.TIM1): // TIM1 clock enable
		stm32.RCC.SetAPBENR2_TIM1EN(1)
	case unsafe.Pointer(stm32.SPI1): // SPI1 clock enable
		stm32.RCC.SetAPBENR2_SPI1EN(1)
	case unsafe.Pointer(stm32.USART1): // USART1 clock enable
		stm32.RCC.SetAPBENR2_USART1EN(1)
	case unsafe.Pointer(stm32.TIM14): // TIM14 clock enable
		stm32.RCC.SetAPBENR2_TIM14EN(1)
	case unsafe.Pointer(stm32.TIM15): // TIM15 clock enable
		stm32.RCC.SetAPBENR2_TIM15EN(1)
	case unsafe.Pointer(stm32.TIM16): // TIM16 clock enable
		stm32.RCC.SetAPBENR2_TIM16EN(1)
	case unsafe.Pointer(stm32.TIM17): // TIM17 clock enable
		stm32.RCC.SetAPBENR2_TIM17EN(1)
	case unsafe.Pointer(stm32.ADC): // ADC clock enable
		stm32.RCC.SetAPBENR2_ADCEN(1)
	case unsafe.Pointer(stm32.FDCAN1), unsafe.Pointer(stm32.FDCAN2): // FDCAN clock enable
		stm32.RCC.SetAPBENR1_FDCANEN(1)
	}
}

//---------- Timer related code

// Alternate function constants for STM32G0
const (
	AF0_SYSTEM                    = 0
	AF1_TIM1_TIM2_TIM3_LPTIM1     = 1
	AF2_TIM1_TIM2_TIM3_TIM14_I2C2 = 2
	AF3_USART5_USART6_LPUART2     = 3
	AF3_FDCAN1_FDCAN2             = 3 // FDCAN on PC2/PC3/PC4/PC5, PD12/PD13/PD14/PD15
	AF4_USART1_USART2_TIM14       = 4
	AF5_SPI1_SPI2_TIM16_TIM17     = 5
	AF6_SPI2_USART3_USART4_I2C1   = 6
	AF7_USART1_USART2_COMP1_COMP2 = 7
	AF8_I2C1_I2C2_UCPD1_UCPD2     = 8
	AF9_SPI2_TIM14_TIM15          = 9
	AF9_FDCAN1_FDCAN2             = 9 // FDCAN on PA11/PA12, PB8/PB9
)

var (
	TIM1 = TIM{
		EnableRegister: &stm32.RCC.APBENR2,
		EnableFlag:     stm32.RCC_APBENR2_TIM1EN,
		Device:         stm32.TIM1,
		Channels: [4]TimerChannel{
			{Pins: []PinFunction{{PA8, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}}},
			{Pins: []PinFunction{{PA9, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}}},
			{Pins: []PinFunction{{PA10, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}}},
			{Pins: []PinFunction{{PA11, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM2 = TIM{
		EnableRegister: &stm32.RCC.APBENR1,
		EnableFlag:     stm32.RCC_APBENR1_TIM2EN,
		Device:         stm32.TIM2,
		Channels: [4]TimerChannel{
			{Pins: []PinFunction{{PA0, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}, {PA5, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}, {PA15, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}}},
			{Pins: []PinFunction{{PA1, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}, {PB3, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}}},
			{Pins: []PinFunction{{PA2, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}, {PB10, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}}},
			{Pins: []PinFunction{{PA3, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}, {PB11, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM3 = TIM{
		EnableRegister: &stm32.RCC.APBENR1,
		EnableFlag:     stm32.RCC_APBENR1_TIM3EN,
		Device:         stm32.TIM3,
		Channels: [4]TimerChannel{
			{Pins: []PinFunction{{PA6, AF1_TIM1_TIM2_TIM3_LPTIM1}, {PB4, AF1_TIM1_TIM2_TIM3_LPTIM1}, {PC6, AF1_TIM1_TIM2_TIM3_LPTIM1}}},
			{Pins: []PinFunction{{PA7, AF1_TIM1_TIM2_TIM3_LPTIM1}, {PB5, AF1_TIM1_TIM2_TIM3_LPTIM1}, {PC7, AF1_TIM1_TIM2_TIM3_LPTIM1}}},
			{Pins: []PinFunction{{PB0, AF1_TIM1_TIM2_TIM3_LPTIM1}, {PC8, AF1_TIM1_TIM2_TIM3_LPTIM1}}},
			{Pins: []PinFunction{{PB1, AF1_TIM1_TIM2_TIM3_LPTIM1}, {PC9, AF1_TIM1_TIM2_TIM3_LPTIM1}}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM6 = TIM{
		EnableRegister: &stm32.RCC.APBENR1,
		EnableFlag:     stm32.RCC_APBENR1_TIM6EN,
		Device:         stm32.TIM6,
		Channels: [4]TimerChannel{
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM7 = TIM{
		EnableRegister: &stm32.RCC.APBENR1,
		EnableFlag:     stm32.RCC_APBENR1_TIM7EN,
		Device:         stm32.TIM7,
		Channels: [4]TimerChannel{
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
		},
		busFreq: APB1_TIM_FREQ,
	}

	TIM14 = TIM{
		EnableRegister: &stm32.RCC.APBENR2,
		EnableFlag:     stm32.RCC_APBENR2_TIM14EN,
		Device:         stm32.TIM14,
		Channels: [4]TimerChannel{
			{Pins: []PinFunction{{PA4, AF4_USART1_USART2_TIM14}, {PA7, AF4_USART1_USART2_TIM14}, {PB1, AF0_SYSTEM}}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM15 = TIM{
		EnableRegister: &stm32.RCC.APBENR2,
		EnableFlag:     stm32.RCC_APBENR2_TIM15EN,
		Device:         stm32.TIM15,
		Channels: [4]TimerChannel{
			{Pins: []PinFunction{{PA2, AF5_SPI1_SPI2_TIM16_TIM17}, {PB14, AF5_SPI1_SPI2_TIM16_TIM17}}},
			{Pins: []PinFunction{{PA3, AF5_SPI1_SPI2_TIM16_TIM17}, {PB15, AF5_SPI1_SPI2_TIM16_TIM17}}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM16 = TIM{
		EnableRegister: &stm32.RCC.APBENR2,
		EnableFlag:     stm32.RCC_APBENR2_TIM16EN,
		Device:         stm32.TIM16,
		Channels: [4]TimerChannel{
			{Pins: []PinFunction{{PA6, AF5_SPI1_SPI2_TIM16_TIM17}, {PB8, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}

	TIM17 = TIM{
		EnableRegister: &stm32.RCC.APBENR2,
		EnableFlag:     stm32.RCC_APBENR2_TIM17EN,
		Device:         stm32.TIM17,
		Channels: [4]TimerChannel{
			{Pins: []PinFunction{{PA7, AF5_SPI1_SPI2_TIM16_TIM17}, {PB9, AF2_TIM1_TIM2_TIM3_TIM14_I2C2}}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
			{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}
)

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(stm32.IRQ_TIM1_BRK_UP_TRG_COM, TIM1.handleUPInterrupt)
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleUPInterrupt)
	case &TIM3:
		return interrupt.New(stm32.IRQ_TIM3_TIM4, TIM3.handleUPInterrupt)
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6_DAC_LPTIM1, TIM6.handleUPInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7, TIM7.handleUPInterrupt)
	case &TIM14:
		return interrupt.New(stm32.IRQ_TIM14, TIM14.handleUPInterrupt)
	case &TIM15:
		return interrupt.New(stm32.IRQ_TIM15, TIM15.handleUPInterrupt)
	case &TIM16:
		return interrupt.New(stm32.IRQ_TIM16, TIM16.handleUPInterrupt)
	case &TIM17:
		return interrupt.New(stm32.IRQ_TIM17, TIM17.handleUPInterrupt)
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
		return interrupt.New(stm32.IRQ_TIM3_TIM4, TIM3.handleOCInterrupt)
	case &TIM6:
		return interrupt.New(stm32.IRQ_TIM6_DAC_LPTIM1, TIM6.handleOCInterrupt)
	case &TIM7:
		return interrupt.New(stm32.IRQ_TIM7, TIM7.handleOCInterrupt)
	case &TIM14:
		return interrupt.New(stm32.IRQ_TIM14, TIM14.handleOCInterrupt)
	case &TIM15:
		return interrupt.New(stm32.IRQ_TIM15, TIM15.handleOCInterrupt)
	case &TIM16:
		return interrupt.New(stm32.IRQ_TIM16, TIM16.handleOCInterrupt)
	case &TIM17:
		return interrupt.New(stm32.IRQ_TIM17, TIM17.handleOCInterrupt)
	}

	return interrupt.Interrupt{}
}

func (t *TIM) enableMainOutput() {
	t.Device.SetBDTR_MOE(1)
}

type arrtype = uint32
type psctype = uint32
type arrRegType = volatile.Register32

const (
	ARR_MAX = 0x10000
	PSC_MAX = 0x10000
)

func initRNG() {
	// STM32G0B1 does not have a hardware RNG peripheral
	// RNG is available on some other STM32G0 variants
}

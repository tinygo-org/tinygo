//go:build stm32wlx
// +build stm32wlx

package machine

// Peripheral abstraction layer for the stm32wle5

import (
	"device/stm32"
	"math/bits"
	"runtime/interrupt"
	"runtime/volatile"
	"unsafe"
)

const (
	AF0_SYSTEM             = 0
	AF1_TIM1_2_LPTIM1      = 1
	AF2_TIM1_2             = 2
	AF3_SPIS2_TIM1_LPTIM3  = 3
	AF4_I2C1_2_3           = 4
	AF5_SPI1_SPI2S2        = 5
	AF6_RF                 = 6
	AF7_USART1_2           = 7
	AF8_LPUART1            = 8
	AF12_COMP1_2_TIM1      = 12
	AF13_DEBUG             = 13
	AF14_TIM2_16_17_LPTIM2 = 14
	AF15_EVENTOUT          = 15
)

const (
	SYSCLK        = 48e6
	APB1_TIM_FREQ = SYSCLK
	APB2_TIM_FREQ = SYSCLK
)

func CPUFrequency() uint32 {
	return SYSCLK
}

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

	PH3 = portH + 3
)

func (p Pin) getPort() *stm32.GPIO_Type {
	switch p / 16 {
	case 0:
		return stm32.GPIOA
	case 1:
		return stm32.GPIOB
	case 2:
		return stm32.GPIOC
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
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOAEN)
	case 1:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOBEN)
	case 2:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOCEN)
	case 7:
		stm32.RCC.AHB2ENR.SetBits(stm32.RCC_AHB2ENR_GPIOHEN)
	default:
		panic("machine: unknown port")
	}
}

// Enable peripheral clock
func enableAltFuncClock(bus unsafe.Pointer) {
	switch bus {
	// APB1ENR1
	case unsafe.Pointer(stm32.LPTIM1): // LPTIM1 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_LPTIM1EN)
	case unsafe.Pointer(stm32.DAC): // DAC clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_DAC1EN)
	case unsafe.Pointer(stm32.I2C3): // I2C3 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_I2C3EN)
	case unsafe.Pointer(stm32.I2C2): // I2C2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_I2C2EN)
	case unsafe.Pointer(stm32.I2C1): // I2C1 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_I2C1EN)
	case unsafe.Pointer(stm32.USART2): // USART2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_USART2EN)
	case unsafe.Pointer(stm32.SPI2): // SPI2S2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_SPI2S2EN)
	case unsafe.Pointer(stm32.WWDG): // Window watchdog clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_WWDGEN)
	case unsafe.Pointer(stm32.TIM2): // TIM2 clock enable
		stm32.RCC.APB1ENR1.SetBits(stm32.RCC_APB1ENR1_TIM2EN)
	// APB1ENR2
	case unsafe.Pointer(stm32.LPTIM3): // LPTIM3 clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_LPTIM3EN)
	case unsafe.Pointer(stm32.LPTIM2): // LPTIM2 clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_LPTIM2EN)
	case unsafe.Pointer(stm32.LPUART): // LPUART clock enable
		stm32.RCC.APB1ENR2.SetBits(stm32.RCC_APB1ENR2_LPUART1EN)
	//APB2ENR
	case unsafe.Pointer(stm32.TIM17): // TIM17 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM17EN)
	case unsafe.Pointer(stm32.TIM16): // TIM16 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM16EN)
	case unsafe.Pointer(stm32.USART1): // USART1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_USART1EN)
	case unsafe.Pointer(stm32.SPI1): // SPI1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_SPI1EN)
	case unsafe.Pointer(stm32.TIM1): // TIM1 clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_TIM1EN)
	case unsafe.Pointer(stm32.ADC): // ADC clock enable
		stm32.RCC.APB2ENR.SetBits(stm32.RCC_APB2ENR_ADCEN)

	}
}

func handlePinInterrupt(pin uint8) {
	if stm32.EXTI.PR1.HasBits(1 << pin) {
		// Writing 1 to the pending register clears the
		// pending flag for that bit
		stm32.EXTI.PR1.Set(1 << pin)

		callback := pinCallbacks[pin]
		if callback != nil {
			callback(interruptPins[pin])
		}
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

// -- SPI ----------------------------------------------------------------------

type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector uint8
}

func (spi SPI) config8Bits() {
	// Set rx threshold to 8-bits, so RXNE flag is set for 1 byte
	// (common STM32 SPI implementation does 8-bit transfers only)
	spi.Bus.CR2.SetBits(stm32.SPI_CR2_FRXTH)
}

func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}

func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var clock uint32

	// We keep this switch and separate management of SPI Clocks
	// for future improvement of system/bus clocks and prescalers
	switch spi.Bus {
	case stm32.SPI1:
		clock = CPUFrequency()
	case stm32.SPI2, stm32.SPI3:
		clock = CPUFrequency()
	}

	// limit requested frequency to bus frequency and min frequency (DIV256)
	freq := config.Frequency
	if min := clock / 256; freq < min {
		freq = min
	} else if freq > clock {
		freq = clock
	}

	// calculate the exact clock divisor (freq=clock/div -> div=clock/freq).
	// truncation is fine, since it produces a less-than-or-equal divisor, and
	// thus a greater-than-or-equal frequency.
	// divisors only come in consecutive powers of 2, so we can use log2 (or,
	// equivalently, bits.Len - 1) to convert to respective enum value.
	div := bits.Len32(clock/freq) - 1

	// but DIV1 (2^0) is not permitted, as the least divisor is DIV2 (2^1), so
	// subtract 1 from the log2 value, keeping a lower bound of 0
	if div < 0 {
		div = 0
	} else if div > 0 {
		div--
	}

	// finally, shift the enumerated value into position for SPI CR1
	return uint32(div) << stm32.SPI_CR1_BR_Pos
}

//---------- I2C related code

// Gets the value for TIMINGR register
func (i2c *I2C) getFreqRange() uint32 {
	// This is a 'magic' value calculated by STM32CubeMX
	// for 48Mhz PCLK1.
	// TODO: Do calculations based on PCLK1
	return 0x20303E5D
}

//---------- UART related code

// Configure the UART.
func (uart UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.TxAltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.RxAltFuncSelector)
}

// UART baudrate calc based on the bus and clockspeed
// NOTE: keep this in sync with the runtime/runtime_stm32wle5.go clock init code
func (uart *UART) getBaudRateDivisor(baudRate uint32) uint32 {
	var br uint32
	uartClock := CPUFrequency() // No Prescaler configuration
	br = uint32((uartClock + baudRate/2) / baudRate)
	return (br)
}

// Register names vary by ST processor, these are for STM L5
func (uart *UART) setRegisters() {
	uart.rxReg = &uart.Bus.RDR
	uart.txReg = &uart.Bus.TDR
	uart.statusReg = &uart.Bus.ISR
	uart.txEmptyFlag = stm32.USART_ISR_TXFNF //(TXFNF == TXE == bit 7, but depends alternate RM0461/1094)
}

//---------- Timer related code

var (
	TIM1 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM1EN,
		Device:         stm32.TIM1,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA8, AF1_TIM1_2_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA9, AF1_TIM1_2_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA10, AF1_TIM1_2_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA11, AF1_TIM1_2_LPTIM1}}},
		},
		busFreq: APB2_TIM_FREQ,
	}
	TIM2 = TIM{
		EnableRegister: &stm32.RCC.APB1ENR1,
		EnableFlag:     stm32.RCC_APB1ENR1_TIM2EN,
		Device:         stm32.TIM2,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA0, AF1_TIM1_2_LPTIM1}, {PA5, AF1_TIM1_2_LPTIM1}, {PA15, AF1_TIM1_2_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA1, AF1_TIM1_2_LPTIM1}, {PB3, AF1_TIM1_2_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA2, AF1_TIM1_2_LPTIM1}, {PB10, AF1_TIM1_2_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{{PA3, AF1_TIM1_2_LPTIM1}, {PB11, AF1_TIM1_2_LPTIM1}}},
		},
		busFreq: APB1_TIM_FREQ,
	}
	TIM16 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM16EN,
		Device:         stm32.TIM16,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA6, AF14_TIM2_16_17_LPTIM2}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}
	TIM17 = TIM{
		EnableRegister: &stm32.RCC.APB2ENR,
		EnableFlag:     stm32.RCC_APB2ENR_TIM17EN,
		Device:         stm32.TIM17,
		Channels: [4]TimerChannel{
			TimerChannel{Pins: []PinFunction{{PA7, AF1_TIM1_2_LPTIM1}, {PB9, AF1_TIM1_2_LPTIM1}}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
			TimerChannel{Pins: []PinFunction{}},
		},
		busFreq: APB2_TIM_FREQ,
	}
)

func (t *TIM) registerUPInterrupt() interrupt.Interrupt {
	switch t {
	case &TIM1:
		return interrupt.New(stm32.IRQ_TIM1_UP, TIM1.handleUPInterrupt)
	case &TIM2:
		return interrupt.New(stm32.IRQ_TIM2, TIM2.handleUPInterrupt)
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
	case &TIM16:
		return interrupt.New(stm32.IRQ_TIM16, TIM16.handleOCInterrupt)
	case &TIM17:
		return interrupt.New(stm32.IRQ_TIM17, TIM17.handleOCInterrupt)
	}

	return interrupt.Interrupt{}
}

func (t *TIM) enableMainOutput() {
	t.Device.BDTR.SetBits(stm32.TIM_BDTR_MOE)
}

func initRNG() {
	stm32.RCC.AHB3ENR.SetBits(stm32.RCC_AHB3ENR_RNGEN)
	stm32.RNG.CR.SetBits(stm32.RNG_CR_RNGEN)
}

//----------

type arrtype = uint32
type arrRegType = volatile.Register32

const (
	ARR_MAX = 0x10000
	PSC_MAX = 0x10000
)

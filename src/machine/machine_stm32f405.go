// +build stm32f405

package machine

// Peripheral abstraction layer for the stm32f405

import (
	"device/stm32"
	"math/bits"
	"runtime/interrupt"
)

func CPUFrequency() uint32 {
	return 168000000
}

// -- UART ---------------------------------------------------------------------

type UART struct {
	Buffer          *RingBuffer
	Bus             *stm32.USART_Type
	Interrupt       interrupt.Interrupt
	AltFuncSelector stm32.AltFunc
}

func (uart UART) configurePins(config UARTConfig) {
	// enable the alternate functions on the TX and RX pins
	config.TX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTTX}, uart.AltFuncSelector)
	config.RX.ConfigureAltFunc(PinConfig{Mode: PinModeUARTRX}, uart.AltFuncSelector)
}

func (uart UART) getBaudRateDivisor(baudRate uint32) uint32 {
	var clock uint32
	switch uart.Bus {
	case stm32.USART1, stm32.USART6:
		clock = CPUFrequency() / 2 // APB2 Frequency
	case stm32.USART2, stm32.USART3, stm32.UART4, stm32.UART5:
		clock = CPUFrequency() / 4 // APB1 Frequency
	}
	return clock / baudRate
}

// -- SPI ----------------------------------------------------------------------

type SPI struct {
	Bus             *stm32.SPI_Type
	AltFuncSelector stm32.AltFunc
}

func (spi SPI) configurePins(config SPIConfig) {
	config.SCK.ConfigureAltFunc(PinConfig{Mode: PinModeSPICLK}, spi.AltFuncSelector)
	config.SDO.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDO}, spi.AltFuncSelector)
	config.SDI.ConfigureAltFunc(PinConfig{Mode: PinModeSPISDI}, spi.AltFuncSelector)
}

func (spi SPI) getBaudRate(config SPIConfig) uint32 {
	var clock uint32
	switch spi.Bus {
	case stm32.SPI1:
		clock = CPUFrequency() / 2
	case stm32.SPI2, stm32.SPI3:
		clock = CPUFrequency() / 4
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

// -- I2C ----------------------------------------------------------------------

type I2C struct {
	Bus             *stm32.I2C_Type
	AltFuncSelector stm32.AltFunc
}

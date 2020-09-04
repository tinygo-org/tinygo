// +build stm32f405

package machine

// Peripheral abstraction layer for the stm32f405

import (
	"device/stm32"
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

func (spi SPI) configurePins(config SPIConfig)      {}
func (spi SPI) getBaudRate(config SPIConfig) uint32 { return 0 }

// -- I2C ----------------------------------------------------------------------

type I2C struct {
	Bus             *stm32.I2C_Type
	AltFuncSelector stm32.AltFunc
}

func (i2c I2C) configurePins(config I2CConfig)      {}
func (i2c I2C) getBaudRate(config I2CConfig) uint32 { return 0 }

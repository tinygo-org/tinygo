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

func (uart UART) configurePins(config UARTConfig)           {}
func (uart UART) getBaudRateDivisor(baudRate uint32) uint32 { return 0 }

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

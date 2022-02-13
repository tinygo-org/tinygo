//go:build bluepill
// +build bluepill

package machine

import (
	"runtime/interrupt"
	"tinygo.org/x/device/stm32"
)

const (
	LED = PC13
)

const (
	// This board does not have a user button, so
	// use first GPIO pin by default
	BUTTON = PA0
)

// Analog Pins
const (
	ADC0 = PA0
	ADC1 = PA1
	ADC2 = PA2
	ADC3 = PA3
	ADC4 = PA4
	ADC5 = PA5
	ADC6 = PA6
	ADC7 = PA7
	ADC8 = PB0
	ADC9 = PB1
)

var DefaultUART = UART1

// UART pins
const (
	UART_TX_PIN     = PA9
	UART_RX_PIN     = PA10
	UART_ALT_TX_PIN = PB6
	UART_ALT_RX_PIN = PB7
)

var (
	// USART1 is the first hardware serial port on the STM32.
	UART1  = &_UART1
	_UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    stm32.USART1,
	}
	UART2  = &_UART2
	_UART2 = UART{
		Buffer: NewRingBuffer(),
		Bus:    stm32.USART2,
	}
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART1, _UART1.handleInterrupt)
	UART2.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART2.handleInterrupt)
}

// SPI pins
const (
	SPI0_SCK_PIN = PA5
	SPI0_SDO_PIN = PA7
	SPI0_SDI_PIN = PA6
)

// I2C pins
const (
	I2C0_SDA_PIN = PB7
	I2C0_SCL_PIN = PB6
)

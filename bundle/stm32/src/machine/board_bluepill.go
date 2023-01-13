//go:build bluepill

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

// Pins printed on the silkscreen
const (
	C13 = PC13
	C14 = PC14
	C15 = PC15
	A0  = PA0
	A1  = PA1
	A2  = PA2
	A3  = PA3
	A4  = PA4
	A5  = PA5
	A6  = PA6
	A7  = PA7
	B0  = PB0
	B1  = PB1
	B10 = PB10
	B11 = PB11
	B12 = PB12
	B13 = PB13
	B14 = PB14
	B15 = PB15
	A8  = PA8
	A9  = PA9
	A10 = PA10
	A11 = PA11
	A12 = PA12
	A13 = PA13
	A14 = PA14
	A15 = PA15
	B3  = PB3
	B4  = PB4
	B5  = PB5
	B6  = PB6
	B7  = PB7
	B8  = PB8
	B9  = PB9
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

const (
	// This board does not have a user button, so
	// use first GPIO pin by default
	BUTTON = PA0

	LED = PC13
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

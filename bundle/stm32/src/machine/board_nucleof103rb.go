//go:build nucleof103rb

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	LED         = LED_BUILTIN
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PA5
)

const (
	BUTTON      = BUTTON_USER
	BUTTON_USER = PC13
)

// UART pins
const (
	UART_TX_PIN     = PA2
	UART_RX_PIN     = PA3
	UART_ALT_TX_PIN = PD5
	UART_ALT_RX_PIN = PD6
)

var (
	// USART2 is the hardware serial port connected to the onboard ST-LINK
	// debugger to be exposed as virtual COM port over USB on Nucleo boards.
	UART2  = &_UART2
	_UART2 = UART{
		Buffer: NewRingBuffer(),
		Bus:    stm32.USART2,
	}
	DefaultUART = UART2
)

func init() {
	UART2.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART2.handleInterrupt)
}

// SPI pins
const (
	SPI0_SCK_PIN = PA5
	SPI0_SDI_PIN = PA6
	SPI0_SDO_PIN = PA7
)

// I2C pins
const (
	I2C0_SCL_PIN = PB6
	I2C0_SDA_PIN = PB7
)

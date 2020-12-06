// +build nucleof722ze

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	LED         = LED_BUILTIN
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PB0
	LED_BLUE    = PB7
	LED_RED     = PB14
)

const (
	BUTTON      = BUTTON_USER
	BUTTON_USER = PC13
)

// UART pins
const (
	// PD8 and PD9 are connected to the ST-Link Virtual Com Port (VCP)
	UART_TX_PIN = PD8
	UART_RX_PIN = PD9
	UART_ALT_FN = 7 // GPIO_AF7_UART3
)

var (
	// USART3 is the hardware serial port connected to the onboard ST-LINK
	// debugger to be exposed as virtual COM port over USB on Nucleo boards.
	// Both UART0 and UART1 refer to USART2.
	UART0 = UART{
		Buffer:          NewRingBuffer(),
		Bus:             stm32.USART3,
		AltFuncSelector: UART_ALT_FN,
	}
	UART1 = &UART0
)

func init() {
	UART0.Interrupt = interrupt.New(stm32.IRQ_USART3, UART0.handleInterrupt)
}

// SPI pins
const (
	SPI0_SCK_PIN = PA5
	SPI0_SDI_PIN = PA6
	SPI0_SDO_PIN = PA7
)

// I2C pins
const (
	SCL_PIN = PB6
	SDA_PIN = PB7
)

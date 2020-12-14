// +build nucleol552ze

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	LED         = LED_BUILTIN
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PC7
	LED_BLUE    = PB7
	LED_RED     = PA9
)

const (
	BUTTON      = BUTTON_USER
	BUTTON_USER = PC13
)

// UART pins
const (
	// PG7 and PG8 are connected to the ST-Link Virtual Com Port (VCP)
	UART_TX_PIN = PG7
	UART_RX_PIN = PG8
	UART_ALT_FN = 8 // GPIO_AF8_LPUART1
)

var (
	// LPUART1 is the hardware serial port connected to the onboard ST-LINK
	// debugger to be exposed as virtual COM port over USB on Nucleo boards.
	// Both UART0 and UART1 refer to LPUART1.
	UART0 = UART{
		Buffer:          NewRingBuffer(),
		Bus:             stm32.LPUART1,
		AltFuncSelector: UART_ALT_FN,
	}
	UART1 = &UART0
)

func init() {
	UART0.Interrupt = interrupt.New(stm32.IRQ_LPUART1, UART0.handleInterrupt)
}

// +build nucleol432kc

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	LED         = LED_BUILTIN
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PB3
)

// UART pins
const (
	// PA2 and PA15 are connected to the ST-Link Virtual Com Port (VCP)
	UART_TX_PIN = PA2
	UART_RX_PIN = PA15
)

// I2C pins
const (
	// PB6 and PB7 are mapped to CN4 pin 7 and CN4 pin 8 respectively with the
	// default solder bridge settings
	I2C0_SCL_PIN = PB6
	I2C0_SDA_PIN = PB7
)

var (
	// USART2 is the hardware serial port connected to the onboard ST-LINK
	// debugger to be exposed as virtual COM port over USB on Nucleo boards.
	// Both UART0 and UART1 refer to USART2.
	UART0 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: 7,
		RxAltFuncSelector: 3,
	}
	UART1 = &UART0
)

func init() {
	UART0.Interrupt = interrupt.New(stm32.IRQ_USART2, UART0.handleInterrupt)
}

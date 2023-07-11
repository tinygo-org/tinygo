//go:build nucleol552ze

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	LED_GREEN   = PC7
	LED_BLUE    = PB7
	LED_RED     = PA9
	LED_BUILTIN = LED_GREEN
	LED         = LED_BUILTIN
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
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.LPUART1,
		TxAltFuncSelector: UART_ALT_FN,
		RxAltFuncSelector: UART_ALT_FN,
	}
	DefaultUART = UART1
)

const (
	I2C0_SCL_PIN = PB8
	I2C0_SDA_PIN = PB9
)

var (
	// I2C1 is documented, alias to I2C0 as well
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: 4,
	}
	I2C0 = I2C1
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_LPUART1, _UART1.handleInterrupt)
}

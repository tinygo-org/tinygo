//go:build nucleof722ze

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
	UART1  = &_UART1
	_UART1 = UART{
		UARTCommon:        NewUARTCommon(),
		Bus:               stm32.USART3,
		TxAltFuncSelector: UART_ALT_FN,
		RxAltFuncSelector: UART_ALT_FN,
	}
	DefaultUART = UART1
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART3, _UART1.handleInterrupt)
}

// SPI pins
const (
	SPI0_SCK_PIN = PA5
	SPI0_SDI_PIN = PA6
	SPI0_SDO_PIN = PA7
)

// I2C pins
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

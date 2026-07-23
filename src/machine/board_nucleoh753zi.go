//go:build nucleoh753zi

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const xtalHz = 8_000_000
const hseBypass = true

const (
	// Arduino Pins
	A0 = PA3
	A1 = PC0
	A2 = PC3
	A3 = PB1
	A4 = PC2
	A5 = PF10

	D0  = PG9
	D1  = PG14
	D2  = PF15
	D3  = PE13
	D4  = PF14
	D5  = PE11
	D6  = PE9
	D7  = PF13
	D8  = PF12
	D9  = PD15
	D10 = PD14
	D11 = PA7
	D12 = PA6
	D13 = PA5
	D14 = PB9
	D15 = PB8
)

const (
	LED         = LED_BUILTIN
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PB0
	LED_YELLOW  = PE1
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
	UART_ALT_FN = AF7_SPI2_3_USART1_2_3_UART5_SPDIFRX
)

var (
	// USART3 is the hardware serial port connected to the onboard ST-LINK
	// debugger to be exposed as virtual COM port over USB on Nucleo boards.
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
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

var (
	SPI1 = &SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: AF5_SPI1_2_3_4_5_6_I2S,
	}
	SPI0 = SPI1
)

// I2C pins
const (
	I2C0_SCL_PIN = PB8
	I2C0_SDA_PIN = PB9
)

var (
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: AF4_I2C1_2_3_4_USART1,
	}
	I2C0 = I2C1
)

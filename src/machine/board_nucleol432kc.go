//go:build nucleol432kc
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

const (
	// This board does not have a user button, so
	// use first GPIO pin by default
	BUTTON = PA0
)

const (
	// Arduino Pins
	A0 = PA0
	A1 = PA1
	A2 = PA3
	A3 = PA4
	A4 = PA5
	A5 = PA6
	A6 = PA7
	A7 = PA2

	D0  = PA10
	D1  = PA9
	D2  = PA12
	D3  = PB0
	D4  = PB7
	D5  = PB6
	D6  = PB1
	D7  = PC14
	D8  = PC15
	D9  = PA8
	D10 = PA11
	D11 = PB5
	D12 = PB4
	D13 = PB3
)

const (
	// UART pins
	// PA2 and PA15 are connected to the ST-Link Virtual Com Port (VCP)
	UART_TX_PIN = PA2
	UART_RX_PIN = PA15

	// I2C pins
	// With default solder bridge settings:
	//    PB6 / Arduino D5 / CN3 Pin 8 is SCL
	//    PB7 / Arduino D4 / CN3 Pin 7 is SDA
	I2C0_SCL_PIN = PB6
	I2C0_SDA_PIN = PB7

	// SPI pins
	SPI1_SCK_PIN = PB3
	SPI1_SDI_PIN = PB5
	SPI1_SDO_PIN = PB4
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN
)

var (
	// USART2 is the hardware serial port connected to the onboard ST-LINK
	// debugger to be exposed as virtual COM port over USB on Nucleo boards.
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: 7,
		RxAltFuncSelector: 3,
	}
	DefaultUART = UART1

	// I2C1 is documented, alias to I2C0 as well
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: 4,
	}
	I2C0 = I2C1

	// SPI1 is documented, alias to SPI0 as well
	SPI1 = &SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: 5,
	}
	SPI0 = SPI1
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART1.handleInterrupt)
}

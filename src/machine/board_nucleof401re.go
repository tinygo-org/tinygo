//go:build nucleof401re

// Schematic: https://www.st.com/resource/en/user_manual/um1724-stm32-nucleo64-boards-mb1136-stmicroelectronics.pdf
// Datasheet: https://www.st.com/resource/en/datasheet/stm32f401re.pdf

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const xtalHz = 8_000_000

const (
	// Arduino Pins
	A0 = PA0
	A1 = PA1
	A2 = PA4
	A3 = PB0
	A4 = PC1
	A5 = PC0

	D0  = PA3
	D1  = PA2
	D2  = PA10
	D3  = PB3
	D4  = PB5
	D5  = PB4
	D6  = PB10
	D7  = PA8
	D8  = PA9
	D9  = PC7
	D10 = PB6
	D11 = PA7
	D12 = PA6
	D13 = PA5
	D14 = PB9
	D15 = PB8
)

// User LD2: the green LED is a user LED connected to Arduino signal D13
// corresponding to STM32 I/O PA5.
const (
	LED         = LED_BUILTIN
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PA5
)

// BUTTON is the user button B1 connected to PC13 (active low).
const BUTTON = PC13

const (
	// UART pins
	// PA2 and PA3 are connected to the ST-Link Virtual Com Port (VCP).
	UART_TX_PIN  = PA2
	UART_RX_PIN  = PA3
	UART1_TX_PIN = UART_TX_PIN
	UART1_RX_PIN = UART_RX_PIN

	// USART1 on Arduino D8 (PA9 = TX) / D2 (PA10 = RX).
	// Use for external UART communication (e.g. with another board).
	UART2_TX_PIN = PA9
	UART2_RX_PIN = PA10

	// I2C pins
	// PB8 / Arduino D15 is SCL, PB9 / Arduino D14 is SDA.
	I2C0_SCL_PIN = PB8
	I2C0_SDA_PIN = PB9

	// SPI pins
	SPI1_SCK_PIN = PA5
	SPI1_SDI_PIN = PA6
	SPI1_SDO_PIN = PA7
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN
)

var (
	// USART2 is the hardware serial port connected to the onboard ST-LINK
	// debugger, exposed as a virtual COM port over USB.
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: AF7_USART1_2_3,
		RxAltFuncSelector: AF7_USART1_2_3,
	}
	DefaultUART = UART1

	// USART1 on PA9 (TX=D8) / PA10 (RX=D2). Use for external communication.
	UART2  = &_UART2
	_UART2 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART1,
		TxAltFuncSelector: AF7_USART1_2_3,
		RxAltFuncSelector: AF7_USART1_2_3,
	}

	// I2C1 is documented; I2C0 is an alias.
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: AF4_I2C1_2_3,
	}
	I2C0 = I2C1

	// SPI1 is documented; SPI0 is an alias.
	SPI1 = &SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: AF5_SPI1_SPI2,
	}
	SPI0 = SPI1
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART1.handleInterrupt)
	UART2.Interrupt = interrupt.New(stm32.IRQ_USART1, _UART2.handleInterrupt)
}

//go:build nucleol476rg

// Schematic: https://www.st.com/resource/en/user_manual/um1724-stm32-nucleo64-boards-mb1136-stmicroelectronics.pdf
// Datasheet: https://www.st.com/resource/en/datasheet/stm32l476je.pdf

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

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

// User LD2: the green LED is a user LED connected to ARDUINO® signal D13 corresponding
// to STM32 I/O PA5 (pin 21) or PB13 (pin 34) depending on the STM32 target.
const (
	LED         = LED_BUILTIN
	LED_BUILTIN = LED_GREEN
	LED_GREEN   = PA5
)

const (
	// This board does not have a user button, so
	// use first GPIO pin by default
	BUTTON = PA0
)

const (
	// UART pins
	// PA2 and PA3 are connected to the ST-Link Virtual Com Port (VCP)
	UART_TX_PIN = PA2
	UART_RX_PIN = PA3

	// I2C pins
	// With default solder bridge settings:
	//    PB8 / Arduino D5 / CN3 Pin 8 is SCL
	//    PB7 / Arduino D4 / CN3 Pin 7 is SDA
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
	// debugger to be exposed as virtual COM port over USB on Nucleo boards.
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: AF7_USART1_2_3,
		RxAltFuncSelector: AF7_USART1_2_3,
	}
	DefaultUART = UART1

	// I2C1 is documented, alias to I2C0 as well
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: AF4_I2C1_2_3,
	}
	I2C0 = I2C1

	// SPI1 is documented, alias to SPI0 as well
	SPI1 = &SPI{
		Bus:             stm32.SPI1,
		AltFuncSelector: AF5_SPI1_2,
	}
	SPI0 = SPI1
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART1.handleInterrupt)
}

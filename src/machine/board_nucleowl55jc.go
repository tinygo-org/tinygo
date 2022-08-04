//go:build nucleowl55jc
// +build nucleowl55jc

package machine

import (
	"runtime/interrupt"

	"tinygo.org/x/device/stm32"
)

const (
	LED_BLUE  = PB15
	LED_GREEN = PB9
	LED_RED   = PB11
	LED       = LED_RED

	BTN1   = PA0
	BTN2   = PA1
	BTN3   = PC6
	BUTTON = BTN1

	// SubGhz (SPI3)
	SPI0_NSS_PIN = PA4
	SPI0_SCK_PIN = PA5
	SPI0_SDO_PIN = PA6
	SPI0_SDI_PIN = PA7

	//MCU USART1
	UART1_TX_PIN = PB6
	UART1_RX_PIN = PB7

	//MCU USART2
	UART2_RX_PIN = PA3
	UART2_TX_PIN = PA2

	// DEFAULT USART
	UART_RX_PIN = UART2_RX_PIN
	UART_TX_PIN = UART2_TX_PIN

	// I2C1 pins
	I2C1_SCL_PIN  = PA9
	I2C1_SDA_PIN  = PA10
	I2C1_ALT_FUNC = 4

	// I2C2 pins
	I2C2_SCL_PIN  = PA12
	I2C2_SDA_PIN  = PA11
	I2C2_ALT_FUNC = 4

	// I2C0 alias for I2C1
	I2C0_SDA_PIN = I2C1_SDA_PIN
	I2C0_SCL_PIN = I2C1_SCL_PIN
)

var (
	// STM32 UART2 is connected to the embedded STLINKV3 Virtual Com Port
	UART0  = &_UART0
	_UART0 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: 7,
		RxAltFuncSelector: 7,
	}

	// UART1 is free
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART1,
		TxAltFuncSelector: 7,
		RxAltFuncSelector: 7,
	}

	DefaultUART = UART0

	// I2C Busses
	I2C1 = &I2C{
		Bus:             stm32.I2C1,
		AltFuncSelector: I2C1_ALT_FUNC,
	}
	I2C2 = &I2C{
		Bus:             stm32.I2C2,
		AltFuncSelector: I2C2_ALT_FUNC,
	}
	I2C0 = I2C1

	// SPI
	SPI3 = SPI{
		Bus: stm32.SPI3,
	}
)

func init() {
	// Enable UARTs Interrupts
	UART0.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART0.handleInterrupt)
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART1, _UART1.handleInterrupt)
}

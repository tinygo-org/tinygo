//go:build lorae5
// +build lorae5

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	// We assume a LED is connected on PB5
	LED = PB5 // Default LED
)

// SubGhz (SPI3)
const (
	SPI0_NSS_PIN = PA4
	SPI0_SCK_PIN = PA5
	SPI0_SDO_PIN = PA6
	SPI0_SDI_PIN = PA7
)

// UARTS
const (
	//MCU USART1
	UART1_TX_PIN = PB6
	UART1_RX_PIN = PB7

	//MCU USART2
	UART2_TX_PIN = PA2
	UART2_RX_PIN = PA3

	// DEFAULT USART
	UART_TX_PIN = UART1_TX_PIN
	UART_RX_PIN = UART1_RX_PIN
)

var (
	// Console UART
	UART0  = &_UART0
	_UART0 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART1,
		TxAltFuncSelector: AF7_USART1_2,
		RxAltFuncSelector: AF7_USART1_2,
	}
	DefaultUART = UART0

	// SPI
	SPI3 = SPI{
		Bus: stm32.SPI3,
	}
)

func init() {
	// Enable UARTs Interrupts
	UART0.Interrupt = interrupt.New(stm32.IRQ_USART1, _UART0.handleInterrupt)
}

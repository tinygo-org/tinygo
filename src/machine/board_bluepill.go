// +build bluepill

package machine

import "device/stm32"

// https://wiki.stm32duino.com/index.php?title=File:Bluepillpinout.gif
const (
	PA0  = portA + 0
	PA1  = portA + 1
	PA2  = portA + 2
	PA3  = portA + 3
	PA4  = portA + 4
	PA5  = portA + 5
	PA6  = portA + 6
	PA7  = portA + 7
	PA8  = portA + 8
	PA9  = portA + 9
	PA10 = portA + 10
	PA11 = portA + 11
	PA12 = portA + 12
	PA13 = portA + 13
	PA14 = portA + 14
	PA15 = portA + 15
	PB0  = portB + 0
	PB1  = portB + 1
	PB2  = portB + 2
	PB3  = portB + 3
	PB4  = portB + 4
	PB5  = portB + 5
	PB6  = portB + 6
	PB7  = portB + 7
	PB8  = portB + 8
	PB9  = portB + 9
	PB10 = portB + 10
	PB11 = portB + 11
	PB12 = portB + 12
	PB13 = portB + 13
	PB14 = portB + 14
	PB15 = portB + 15
	PC13 = portC + 13
	PC14 = portC + 14
	PC15 = portC + 15
)

const (
	LED = PC13
)

// UART pins
const (
	UART_TX_PIN     = PA9
	UART_RX_PIN     = PA10
	UART_ALT_TX_PIN = PB6
	UART_ALT_RX_PIN = PB7
)

var (
	// USART1 is the first hardware serial port on the STM32.
	// Both UART0 and UART1 refer to USART1.
	UART0 = UART{
		Buffer: NewRingBuffer(),
		Bus:    stm32.USART1,
		IRQVal: stm32.IRQ_USART1,
	}
	UART1 = &UART0
)

//go:export USART1_IRQHandler
func handleUART1() {
	UART1.Receive(byte((UART1.Bus.DR.Get() & 0xFF)))
}

// SPI pins
const (
	SPI0_SCK_PIN  = PA5
	SPI0_MOSI_PIN = PA7
	SPI0_MISO_PIN = PA6
)

// I2C pins
const (
	SDA_PIN = PB7
	SCL_PIN = PB6
)

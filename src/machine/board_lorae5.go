//go:build lorae5

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	// We assume a LED is connected on PB5
	LED = PB5 // Default LED

	// Set the POWER_EN3V3 pin to high to turn
	// on the 3.3V power for all peripherals
	POWER_EN3V3 = PA9

	// Set the POWER_EN5V pin to high to turn
	// on the 5V bus power for all peripherals
	POWER_EN5V = PB10
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
	// MCU USART1
	UART1_TX_PIN = PB6
	UART1_RX_PIN = PB7

	// MCU USART2
	UART2_TX_PIN = PA2
	UART2_RX_PIN = PA3

	// DEFAULT USART
	UART_TX_PIN = UART1_TX_PIN
	UART_RX_PIN = UART1_RX_PIN

	// I2C2 pins
	I2C2_SCL_PIN  = PB15
	I2C2_SDA_PIN  = PA15
	I2C2_ALT_FUNC = 4

	// I2C0 alias for I2C2
	I2C0_SDA_PIN = I2C2_SDA_PIN
	I2C0_SCL_PIN = I2C2_SCL_PIN
)

var (
	// Console UART
	UART0  = &_UART0
	_UART0 = UART{
		UARTCommon:        NewUARTCommon(),
		Bus:               stm32.USART1,
		TxAltFuncSelector: AF7_USART1_2,
		RxAltFuncSelector: AF7_USART1_2,
	}
	DefaultUART = UART0

	// Since we treat UART1 as zero, let's also call it by the real name
	UART1 = UART0

	// UART2
	UART2  = &_UART2
	_UART2 = UART{
		UARTCommon:        NewUARTCommon(),
		Bus:               stm32.USART2,
		TxAltFuncSelector: AF7_USART1_2,
		RxAltFuncSelector: AF7_USART1_2,
	}

	// I2C Busses
	I2C2 = &I2C{
		Bus:             stm32.I2C2,
		AltFuncSelector: I2C2_ALT_FUNC,
	}

	// Set "default" I2C bus to I2C2
	I2C0 = I2C2

	// SPI
	SPI3 = SPI{
		Bus: stm32.SPI3,
	}
)

func init() {
	// Enable UARTs Interrupts
	UART0.Interrupt = interrupt.New(stm32.IRQ_USART1, _UART0.handleInterrupt)
	UART2.Interrupt = interrupt.New(stm32.IRQ_USART2, _UART2.handleInterrupt)
}

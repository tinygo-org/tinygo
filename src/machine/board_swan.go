//go:build swan
// +build swan

package machine

import (
	"device/stm32"
	"runtime/interrupt"
)

const (
	// LED on the SWAN
	LED = PE2

	// UART pins
	//    PA9 and PA10 are connected to the SWAN Tx/Rx
	UART_TX_PIN = PA9
	UART_RX_PIN = PA10

	// I2C pins
	//    PB6 is SCL
	//    PB7 is SDA
	I2C0_SCL_PIN = PB6
	I2C0_SDA_PIN = PB7

	// SPI pins
	SPI1_SCK_PIN = PD1
	SPI1_SDI_PIN = PB14
	SPI1_SDO_PIN = PB15
	SPI0_SCK_PIN = SPI1_SCK_PIN
	SPI0_SDI_PIN = SPI1_SDI_PIN
	SPI0_SDO_PIN = SPI1_SDO_PIN
)

var (
	// USART1 is connected to the TX/RX pins
	UART1  = &_UART1
	_UART1 = UART{
		Buffer:            NewRingBuffer(),
		Bus:               stm32.USART1,
		TxAltFuncSelector: 7,
		RxAltFuncSelector: 7,
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
		Bus:             stm32.SPI2,
		AltFuncSelector: 5,
	}
	SPI0 = SPI1
)

func init() {
	UART1.Interrupt = interrupt.New(stm32.IRQ_USART1, _UART1.handleInterrupt)
}

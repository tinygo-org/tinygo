// +build grandcentral_m4

package machine

import (
	"device/sam"
	"runtime/interrupt"
)

func init() {
	UART1.Interrupt = interrupt.New(sam.IRQ_SERCOM0_2, _UART1.handleInterrupt)
	UART2.Interrupt = interrupt.New(sam.IRQ_SERCOM4_2, _UART2.handleInterrupt)
	UART3.Interrupt = interrupt.New(sam.IRQ_SERCOM1_2, _UART3.handleInterrupt)
	UART4.Interrupt = interrupt.New(sam.IRQ_SERCOM5_2, _UART4.handleInterrupt)
}

// UART on the Grand Central M4
var (
	UART1  = &_UART1
	_UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM0_USART_INT,
		SERCOM: 0,
	}
	UART2  = &_UART2
	_UART2 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM4_USART_INT,
		SERCOM: 4,
	}
	UART3  = &_UART3
	_UART3 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM1_USART_INT,
		SERCOM: 1,
	}
	UART4  = &_UART4
	_UART4 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM5_USART_INT,
		SERCOM: 5,
	}
)

// I2C on the Grand Central M4
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM3_I2CM,
		SERCOM: 3,
	}
	I2C1 = &I2C{
		Bus:    sam.SERCOM6_I2CM,
		SERCOM: 6,
	}
)

// SPI on the Grand Central M4
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM7_SPIM,
		SERCOM: 7,
	}
	SPI1 = SPI{ // SD card
		Bus:    sam.SERCOM2_SPIM,
		SERCOM: 2,
	}
)

var (
	DefaultUART = UART1
)

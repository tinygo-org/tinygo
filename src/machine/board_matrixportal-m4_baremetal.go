// +build sam,atsamd51,matrixportal_m4

package machine

import (
	"device/sam"
	"runtime/interrupt"
)

// UART on the MatrixPortal M4
var (
	Serial = UART1
	UART1  = &_UART1
	_UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM1_USART_INT,
		SERCOM: 1,
	}

	UART2  = &_UART2
	_UART2 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM4_USART_INT,
		SERCOM: 4,
	}
)

func init() {
	UART1.Interrupt = interrupt.New(sam.IRQ_SERCOM1_1, _UART1.handleInterrupt)
	UART2.Interrupt = interrupt.New(sam.IRQ_SERCOM4_1, _UART2.handleInterrupt)
}

// I2C on the MatrixPortal M4
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM5_I2CM,
		SERCOM: 5,
	}
)

// SPI on the MatrixPortal M4
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM3_SPIM,
		SERCOM: 3, // BUG: SDO on SERCOM1!
	}
	NINA_SPI = SPI0

	SPI1 = SPI{
		Bus:    sam.SERCOM0_SPIM,
		SERCOM: 0,
	}
)

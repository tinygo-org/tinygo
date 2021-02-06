// +build sam,atsamd51,metro_m4_airlift

package machine

import (
	"device/sam"
	"runtime/interrupt"
)

var (
	UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM3_USART_INT,
		SERCOM: 3,
	}

	UART2 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM0_USART_INT,
		SERCOM: 0,
	}
)

func init() {
	UART1.Interrupt = interrupt.New(sam.IRQ_SERCOM3_2, UART1.handleInterrupt)
	UART2.Interrupt = interrupt.New(sam.IRQ_SERCOM0_2, UART2.handleInterrupt)
}

// I2C on the Metro M4.
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM5_I2CM,
		SERCOM: 5,
	}
)

// SPI on the Metro M4.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM2_SPIM,
		SERCOM: 2,
	}
	NINA_SPI = SPI0
)

// SPI1 on the Metro M4 on pins 11,12,13
var (
	SPI1 = SPI{
		Bus:    sam.SERCOM1_SPIM,
		SERCOM: 1,
	}
)

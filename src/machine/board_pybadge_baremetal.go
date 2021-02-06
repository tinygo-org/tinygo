// +build sam,atsamd51,pybadge

package machine

import (
	"device/sam"
	"runtime/interrupt"
)

var (
	UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM5_USART_INT,
		SERCOM: 5,
	}

	UART2 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM0_USART_INT,
		SERCOM: 0,
	}
)

func init() {
	UART1.Interrupt = interrupt.New(sam.IRQ_SERCOM5_2, UART1.handleInterrupt)
	UART2.Interrupt = interrupt.New(sam.IRQ_SERCOM0_2, UART2.handleInterrupt)
}

// I2C on the ItsyBitsy M4.
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM2_I2CM,
		SERCOM: 2,
	}
)

// SPI on the PyBadge.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM1_SPIM,
		SERCOM: 1,
	}
)

// TFT SPI on the PyBadge.
var (
	SPI1 = SPI{
		Bus:    sam.SERCOM4_SPIM,
		SERCOM: 4,
	}
)

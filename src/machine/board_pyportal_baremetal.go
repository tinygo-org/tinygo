// +build sam,atsamd51,pyportal

package machine

import (
	"device/sam"
	"runtime/interrupt"
)

var (
	UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM4_USART_INT,
		SERCOM: 4,
	}
)

func init() {
	UART1.Interrupt = interrupt.New(sam.IRQ_SERCOM4_2, UART1.handleInterrupt)
}

// I2C on the PyPortal.
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM5_I2CM,
		SERCOM: 5,
	}
)

// SPI on the PyPortal.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM2_SPIM,
		SERCOM: 2,
	}
	NINA_SPI = SPI0
)

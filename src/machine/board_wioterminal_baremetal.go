// +build sam,atsamd51,wioterminal

package machine

import (
	"device/sam"
	"runtime/interrupt"
)

var (
	UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM2_USART_INT,
		SERCOM: 2,
	}

	// RTL8720D
	UART2 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM1_USART_INT,
		SERCOM: 1,
	}
)

func init() {
	UART1.Interrupt = interrupt.New(sam.IRQ_SERCOM2_2, UART1.handleInterrupt)
	UART2.Interrupt = interrupt.New(sam.IRQ_SERCOM1_2, UART2.handleInterrupt)
}

// I2C on the Wio Terminal
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM4_I2CM,
		SERCOM: 4,
	}

	I2C1 = &I2C{
		Bus:    sam.SERCOM4_I2CM,
		SERCOM: 4,
	}
)

// SPI on the Wio Terminal
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM5_SPIM,
		SERCOM: 5,
	}

	// RTL8720D
	SPI1 = SPI{
		Bus:    sam.SERCOM0_SPIM,
		SERCOM: 0,
	}

	// SD
	SPI2 = SPI{
		Bus:    sam.SERCOM6_SPIM,
		SERCOM: 6,
	}

	// LCD
	SPI3 = SPI{
		Bus:    sam.SERCOM7_SPIM,
		SERCOM: 7,
	}
)

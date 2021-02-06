// +build sam,atsamd21,p1am_100

package machine

import (
	"device/sam"
	"runtime/interrupt"
)

// UART1 on the P1AM-100 connects to the normal TX/RX pins.
var (
	UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM3_USART,
		SERCOM: 5,
	}
)

func init() {
	UART1.Interrupt = interrupt.New(sam.IRQ_SERCOM5, UART1.handleInterrupt)
}

// I2C on the P1AM-100.
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM0_I2CM,
		SERCOM: 0,
	}
)

// SPI on the P1AM-100 is used for Base Controller.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM1_SPI,
		SERCOM: 1,
	}
	BASE_CONTROLLER_SPI = SPI0
)

// SPI1 is connected to the SD card slot on the P1AM-100
var (
	SPI1 = SPI{
		Bus:    sam.SERCOM2_SPI,
		SERCOM: 2,
	}
	SDCARD_SPI = SPI1
)

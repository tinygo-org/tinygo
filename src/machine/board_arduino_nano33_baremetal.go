// +build sam,atsamd21,arduino_nano33

package machine

import (
	"device/sam"
	"runtime/interrupt"
)

// UART1 on the Arduino Nano 33 connects to the onboard NINA-W102 WiFi chip.
var (
	UART1 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM3_USART,
		SERCOM: 3,
	}
)

// UART2 on the Arduino Nano 33 connects to the normal TX/RX pins.
var (
	UART2 = UART{
		Buffer: NewRingBuffer(),
		Bus:    sam.SERCOM5_USART,
		SERCOM: 5,
	}
)

func init() {
	UART1.Interrupt = interrupt.New(sam.IRQ_SERCOM3, UART1.handleInterrupt)
	UART2.Interrupt = interrupt.New(sam.IRQ_SERCOM5, UART2.handleInterrupt)
}

// I2C on the Arduino Nano 33.
var (
	I2C0 = &I2C{
		Bus:    sam.SERCOM4_I2CM,
		SERCOM: 4,
	}
)

// SPI on the Arduino Nano 33.
var (
	SPI0 = SPI{
		Bus:    sam.SERCOM1_SPI,
		SERCOM: 1,
	}
)

// SPI1 is connected to the NINA-W102 chip on the Arduino Nano 33.
var (
	SPI1 = SPI{
		Bus:    sam.SERCOM2_SPI,
		SERCOM: 2,
	}
	NINA_SPI = SPI1
)

// I2S on the Arduino Nano 33.
var (
	I2S0 = I2S{Bus: sam.I2S}
)

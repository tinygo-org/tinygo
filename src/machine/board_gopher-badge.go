//go:build gopher_badger

// This contains the pin mappings for the Gopher Badge.
//
// For more information, see: https://tinygo.org/gopherbadge
package machine

import (
	"device/rp"
	"runtime/interrupt"
)

const (
	NEOPIXELS Pin = GPIO26

	EPD_BUSY_PIN  Pin = GPIO26
	EPD_RESET_PIN Pin = GPIO21
	EPD_DC_PIN    Pin = GPIO20
	EPD_CS_PIN    Pin = GPIO17
	EPD_SCK_PIN   Pin = GPIO18
	EPD_SDO_PIN   Pin = GPIO19

	VBUS_DETECT Pin = GPIO24
	BATTERY     Pin = GPIO29
	ENABLE_3V3  Pin = GPIO10
)

// I2C pins
const (
	I2C0_SDA_PIN Pin = GPIO4
	I2C0_SCL_PIN Pin = GPIO5

	I2C1_SDA_PIN Pin = NoPin
	I2C1_SCL_PIN Pin = NoPin
)

// SPI pins.
const (
	SPI0_SCK_PIN Pin = GPIO18
	SPI0_SDO_PIN Pin = GPIO19
	SPI0_SDI_PIN Pin = GPIO16

	SPI1_SCK_PIN Pin = NoPin
	SPI1_SDO_PIN Pin = NoPin
	SPI1_SDI_PIN Pin = NoPin
)

// QSPI pins¿?
const (
/*
	TODO

SPI0_SD0_PIN Pin = QSPI_SD0
SPI0_SD1_PIN Pin = QSPI_SD1
SPI0_SD2_PIN Pin = QSPI_SD2
SPI0_SD3_PIN Pin = QSPI_SD3
SPI0_SCK_PIN Pin = QSPI_SCLKGPIO6
SPI0_CS_PIN  Pin = QSPI_CS
*/
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Badger 2040"
	usb_STRING_MANUFACTURER = "Pimoroni"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x0003
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

// UART on the RP2040
var (
	UART0  = &_UART0
	_UART0 = UART{
		Buffer: NewRingBuffer(),
		Bus:    rp.UART0,
	}
)

var DefaultUART = UART0

func init() {
	UART0.Interrupt = interrupt.New(rp.IRQ_UART0_IRQ, _UART0.handleInterrupt)
}

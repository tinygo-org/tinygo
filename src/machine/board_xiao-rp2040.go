//go:build xiao_rp2040
// +build xiao_rp2040

// This file contains the pin mappings for the Seeed XIAO RP2040 boards.
//
// XIAO RP2040 is a microcontroller using the Raspberry Pi RP2040 chip.
//
// - https://wiki.seeedstudio.com/XIAO-RP2040/
package machine

import (
	"runtime/interrupt"

	"tinygo.org/x/device/rp"
)

// Digital Pins
const (
	D0  Pin = GPIO26
	D1  Pin = GPIO27
	D2  Pin = GPIO28
	D3  Pin = GPIO29
	D4  Pin = GPIO6
	D5  Pin = GPIO7
	D6  Pin = GPIO0
	D7  Pin = GPIO1
	D8  Pin = GPIO2
	D9  Pin = GPIO3
	D10 Pin = GPIO4
)

// Analog pins
const (
	A0 Pin = D0
	A1 Pin = D1
	A2 Pin = D2
	A3 Pin = D3
)

// Onboard LEDs
const (
	NEOPIXEL = GPIO12

	LED       = GPIO17
	LED_RED   = GPIO17
	LED_GREEN = GPIO16
	LED_BLUE  = GPIO25
)

// I2C pins
const (
	I2C0_SDA_PIN Pin = D4
	I2C0_SCL_PIN Pin = D5

	I2C1_SDA_PIN Pin = NoPin
	I2C1_SCL_PIN Pin = NoPin
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = D8
	SPI0_SDO_PIN Pin = D9
	SPI0_SDI_PIN Pin = D10

	SPI1_SCK_PIN Pin = NoPin
	SPI1_SDO_PIN Pin = NoPin
	SPI1_SDI_PIN Pin = NoPin
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
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

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "XIAO RP2040"
	usb_STRING_MANUFACTURER = "Seeed"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x000a
)

//go:build waveshare_rp2040_zero

// This file contains the pin mappings for the Waveshare RP2040-Zero boards.
//
// Waveshare RP2040-Zero is a microcontroller using the Raspberry Pi RP2040 chip.
//
// - https://www.waveshare.com/wiki/RP2040-Zero
package machine

import (
	"device/rp"
	"runtime/interrupt"
)

// Digital Pins
const (
	D0  Pin = GPIO0
	D1  Pin = GPIO1
	D2  Pin = GPIO2
	D3  Pin = GPIO3
	D4  Pin = GPIO4
	D5  Pin = GPIO5
	D6  Pin = GPIO6
	D7  Pin = GPIO7
	D8  Pin = GPIO8
	D9  Pin = GPIO9
	D10 Pin = GPIO10
	D11 Pin = GPIO11
	D12 Pin = GPIO12
	D13 Pin = GPIO13
	D14 Pin = GPIO14
	D15 Pin = GPIO15
	D16 Pin = GPIO16
	D17 Pin = GPIO17
	D18 Pin = GPIO18
	D19 Pin = GPIO19
	D20 Pin = GPIO20
	D21 Pin = GPIO21
	D22 Pin = GPIO22
	D23 Pin = GPIO23
	D24 Pin = GPIO24
	D25 Pin = GPIO25
	D26 Pin = GPIO26
	D27 Pin = GPIO27
	D28 Pin = GPIO28
	D29 Pin = GPIO29
)

// Analog pins
const (
	A0 Pin = D26
	A1 Pin = D27
	A2 Pin = D28
	A3 Pin = D29
)

// Onboard LEDs
const (
	NEOPIXEL = GPIO16
)

// I2C pins
const (
	I2C0_SDA_PIN Pin = D0
	I2C0_SCL_PIN Pin = D1

	I2C1_SDA_PIN Pin = D2
	I2C1_SCL_PIN Pin = D3
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = D6
	SPI0_SDO_PIN Pin = D3
	SPI0_SDI_PIN Pin = D4

	SPI1_SCK_PIN Pin = D10
	SPI1_SDO_PIN Pin = D11
	SPI1_SDI_PIN Pin = D12
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
	UART1_TX_PIN = GPIO8
	UART1_RX_PIN = GPIO9
)

// UART on the RP2040
var (
	UART0  = &_UART0
	_UART0 = UART{
		UARTCommon: NewUARTCommon(),
		Bus:        rp.UART0,
	}
	UART1  = &_UART1
	_UART1 = UART{
		UARTCommon: NewUARTCommon(),
		Bus:        rp.UART1,
	}
)

var DefaultUART = UART0

func init() {
	UART0.Interrupt = interrupt.New(rp.IRQ_UART0_IRQ, _UART0.handleInterrupt)
	UART1.Interrupt = interrupt.New(rp.IRQ_UART1_IRQ, _UART1.handleInterrupt)
}

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "RP2040-Zero"
	usb_STRING_MANUFACTURER = "Waveshare"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x0003
)

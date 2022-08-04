//go:build macropad_rp2040
// +build macropad_rp2040

package machine

import (
	"runtime/interrupt"

	"tinygo.org/x/device/rp"
)

const (
	NeopixelCount = 12

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

const (
	SWITCH = GPIO0

	KEY1  = GPIO1
	KEY2  = GPIO2
	KEY3  = GPIO3
	KEY4  = GPIO4
	KEY5  = GPIO5
	KEY6  = GPIO6
	KEY7  = GPIO7
	KEY8  = GPIO8
	KEY9  = GPIO9
	KEY10 = GPIO10
	KEY11 = GPIO11
	KEY12 = GPIO12

	LED = GPIO13

	SPEAKER_ENABLE = GPIO14
	SPEAKER        = GPIO16

	ROT_A = GPIO18
	ROT_B = GPIO17

	OLED_CS  = GPIO22
	OLED_RST = GPIO23
	OLED_DC  = GPIO24

	NEOPIXEL = GPIO19
	WS2812   = NEOPIXEL
)

// I2C Default pins on Raspberry Pico.
const (
	I2C0_SDA_PIN = GPIO20
	I2C0_SCL_PIN = GPIO21

	I2C1_SDA_PIN = 31 // not pinned out
	I2C1_SCL_PIN = 31 // not pinned out
)

// SPI default pins
const (
	// Default Serial Clock Bus 1 for SPI communications
	SPI1_SCK_PIN = GPIO26
	// Default Serial Out Bus 1 for SPI communications
	SPI1_SDO_PIN = GPIO27 // Tx
	// Default Serial In Bus 1 for SPI communications
	SPI1_SDI_PIN = GPIO28 // Rx

	SPI0_SCK_PIN = 31 // not pinned out
	SPI0_SDO_PIN = 31 // not pinned out
	SPI0_SDI_PIN = 31 // not pinned out
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

// USB identifiers
const (
	usb_STRING_PRODUCT      = "MacroPad RP2040"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x8107
)

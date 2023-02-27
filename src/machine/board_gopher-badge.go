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
	ADC0  Pin = GPIO26
	ADC1  Pin = GPIO27
	ADC2  Pin = GPIO28
	GPIO4 Pin = GPIO4
	GPIO5 Pin = GPIO5
	GPIO6 Pin = GPIO6
	GPIO7 Pin = GPIO7
	GPIO8 Pin = GPIO8
	GPIO9 Pin = GPIO9

	LED       Pin = GPIO2
	NEOPIXELS Pin = GPIO26
	WS2812    Pin = GPIO26

	BUTTON_A     Pin = GPIO10
	BUTTON_B     Pin = GPIO11
	BUTTON_LEFT  Pin = GPIO25
	BUTTON_UP    Pin = GPIO24
	BUTTON_RIGHT Pin = GPIO22
	BUTTON_DOWN  Pin = GPIO23

	TFT_RST       Pin = GPIO21
	TFT_SDI       Pin = GPIO19
	TFT_SDO       Pin = GPIO16
	TFT_CS        Pin = GPIO17
	TFT_SCL       Pin = GPIO18
	TFT_WRX       Pin = GPIO20
	TFT_BACKLIGHT Pin = GPIO12

	SPEAKER        Pin = GPIO14
	SPEAKER_ENABLE Pin = GPIO3
)

// I2C pins
const (
	I2C0_SDA_PIN Pin = GPIO0
	I2C0_SCL_PIN Pin = GPIO1

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

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Gopher Badger"
	usb_STRING_MANUFACTURER = "TinyGo"
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

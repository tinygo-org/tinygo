//go:build trinkey_qt2040
// +build trinkey_qt2040

// This file contains the pin mappings for the Adafruit Trinkey QT2040 board.
//
// The Trinkey QT2040 is a small development board based on the RP2040 which
// plugs into a USB A port. The board has a minimal pinout: an integrated
// NeoPixel LED and a STEMMA QT I2C port.
//
// - Product:    https://www.adafruit.com/product/5056
// - Overview:   https://learn.adafruit.com/adafruit-trinkey-qt2040
// - Pinouts:    https://learn.adafruit.com/adafruit-trinkey-qt2040/pinouts
// - Datasheets: https://learn.adafruit.com/adafruit-trinkey-qt2040/downloads

package machine

// Onboard crystal oscillator frequency, in MHz
const xoscFreq = 12 // MHz

// Onboard LEDs
const (
	NEOPIXEL = GPIO27
	WS2812   = NEOPIXEL
)

// I2C pins
const (
	I2C0_SDA_PIN = GPIO16
	I2C0_SCL_PIN = GPIO17

	I2C1_SDA_PIN = NoPin
	I2C1_SCL_PIN = NoPin
)

// SPI pins
const (
	SPI0_SCK_PIN = NoPin
	SPI0_SDO_PIN = NoPin
	SPI0_SDI_PIN = NoPin

	SPI1_SCK_PIN = NoPin
	SPI1_SDO_PIN = NoPin
	SPI1_SDI_PIN = NoPin
)

// UART pins
const (
	UART0_TX_PIN = NoPin
	UART0_RX_PIN = NoPin
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

// USB identifiers
const (
	usb_STRING_PRODUCT      = "Trinkey QT2040"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239a
	usb_PID uint16 = 0x8109
)

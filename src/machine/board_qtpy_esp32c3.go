//go:build qtpy_esp32c3

// This file contains the pin mappings for the Adafruit QtPy ESP32C3 boards.
//
// https://learn.adafruit.com/adafruit-qt-py-esp32-c3-wifi-dev-board/pinouts
package machine

// Digital Pins
const (
	D0 = GPIO4
	D1 = GPIO3
	D2 = GPIO1
	D3 = GPIO0
)

// Analog pins (ADC1)
const (
	A0 = GPIO4
	A1 = GPIO3
	A2 = GPIO1
	A3 = GPIO0
)

// UART pins
const (
	RX_PIN = GPIO20
	TX_PIN = GPIO21

	UART_RX_PIN = RX_PIN
	UART_TX_PIN = TX_PIN
)

// I2C pins
const (
	SDA_PIN = GPIO5
	SCL_PIN = GPIO6

	I2C0_SDA_PIN = SDA_PIN
	I2C0_SCL_PIN = SCL_PIN
)

// SPI pins
const (
	SCK_PIN = GPIO10
	MI_PIN  = GPIO8
	MO_PIN  = GPIO7

	SPI_SCK_PIN = SCK_PIN
	SPI_SDI_PIN = MI_PIN
	SPI_SDO_PIN = MO_PIN
)

const (
	NEOPIXEL = GPIO2
	WS2812   = GPIO2

	// also used for boot button.
	// set it to be an input-with-pullup
	BUTTON = GPIO9
)

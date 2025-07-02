//go:build esp32c3_supermini

// This file contains the pin mappings for the ESP32 supermini boards.
//
// - https://web.archive.org/web/20240805232453/https://dl.artronshop.co.th/ESP32-C3%20SuperMini%20datasheet.pdf

package machine

// Digital Pins
const (
	IO0  = GPIO0
	IO1  = GPIO1
	IO2  = GPIO2
	IO3  = GPIO3
	IO4  = GPIO4
	IO5  = GPIO5
	IO6  = GPIO6
	IO7  = GPIO7
	IO8  = GPIO8
	IO9  = GPIO9
	IO10 = GPIO10
	IO20 = GPIO20
	IO21 = GPIO21
)

// Built-in LED
const LED = GPIO8

// Analog pins
const (
	A0 = GPIO0
	A1 = GPIO1
	A2 = GPIO2
	A3 = GPIO3
	A4 = GPIO4
	A5 = GPIO5
)

// UART pins
const (
	UART_RX_PIN = GPIO20
	UART_TX_PIN = GPIO21
)

// I2C pins
const (
	SDA_PIN = GPIO8
	SCL_PIN = GPIO9
)

// SPI pins
const (
	SPI_MISO_PIN = GPIO5
	SPI_MOSI_PIN = GPIO6
	SPI_SS_PIN   = GPIO7
	SPI_SCK_PIN  = GPIO4
)

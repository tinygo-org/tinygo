//go:build xiao_esp32c3
// +build xiao_esp32c3

// This file contains the pin mappings for the Seeed XIAO ESP32C3 boards.
//
// Seeed Studio XIAO ESP32C3 is an IoT mini development board based on
// the Espressif ESP32-C3 WiFi/Bluetooth dual-mode chip.
//
// - https://www.seeedstudio.com/Seeed-XIAO-ESP32C3-p-5431.html
// - https://wiki.seeedstudio.com/XIAO_ESP32C3_Getting_Started/

package machine

// Digital Pins
const (
	D0  = GPIO2
	D1  = GPIO3
	D2  = GPIO4
	D3  = GPIO5
	D4  = GPIO6
	D5  = GPIO7
	D6  = GPIO21
	D7  = GPIO20
	D8  = GPIO8
	D9  = GPIO9
	D10 = GPIO10
)

// Analog pins
const (
	A0 = GPIO2
	A1 = GPIO3
	A2 = GPIO4
	A3 = GPIO5
)

// UART pins
const (
	UART_RX_PIN = GPIO20
	UART_TX_PIN = GPIO21
)

// I2C pins
const (
	SDA_PIN = GPIO6
	SCL_PIN = GPIO7
)

// SPI pins
const (
	SPI_SCK_PIN = GPIO8
	SPI_SDI_PIN = GPIO9
	SPI_SDO_PIN = GPIO10
)

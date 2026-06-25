//go:build xiao_esp32c6

// This file contains the pin mappings for the Seeed XIAO ESP32C6 board.
//
// Seeed Studio XIAO ESP32C6 is an IoT mini development board based on
// the Espressif ESP32-C6 WiFi 6 / Bluetooth 5 / 802.15.4 chip.
//
// - https://www.seeedstudio.com/Seeed-Studio-XIAO-ESP32C6-p-5884.html
// - https://wiki.seeedstudio.com/xiao_esp32c6_getting_started/

package machine

// Digital Pins
const (
	D0  = GPIO0
	D1  = GPIO1
	D2  = GPIO2
	D3  = GPIO21
	D4  = GPIO22
	D5  = GPIO23
	D6  = GPIO16
	D7  = GPIO17
	D8  = GPIO19
	D9  = GPIO20
	D10 = GPIO18
)

// Analog pins
const (
	A0 = GPIO0
	A1 = GPIO1
	A2 = GPIO2
)

// UART pins
const (
	UART_TX_PIN = GPIO16
	UART_RX_PIN = GPIO17
)

// I2C pins
const (
	SDA_PIN = GPIO22
	SCL_PIN = GPIO23
)

// SPI pins
const (
	SPI_SCK_PIN = GPIO19
	SPI_SDI_PIN = GPIO20
	SPI_SDO_PIN = GPIO18
)

// Onboard LEDs
const (
	LED = GPIO15
)

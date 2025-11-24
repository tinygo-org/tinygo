//go:build xiao_esp32s3

// This file contains the pin mappings for the Seeed XIAO ESP32S3 boards.
//
// Seeed Studio XIAO ESP32S3 is an IoT mini development board based on
// the Espressif ESP32-S3 WiFi/Bluetooth dual-mode chip.
//
// - https://www.seeedstudio.com/XIAO-ESP32S3-p-5627.html
// - https://wiki.seeedstudio.com/xiao_esp32s3_getting_started/

package machine

// Digital Pins
const (
	D0  = GPIO1
	D1  = GPIO2
	D2  = GPIO3
	D3  = GPIO4
	D4  = GPIO5
	D5  = GPIO6
	D6  = GPIO43
	D7  = GPIO44
	D8  = GPIO7
	D9  = GPIO8
	D10 = GPIO9
)

// Analog pins
const (
	A0 = GPIO1
	A1 = GPIO2
	A2 = GPIO3
	A3 = GPIO4
)

// UART pins
const (
	UART_RX_PIN = GPIO44
	UART_TX_PIN = GPIO43
)

// I2C pins
const (
	SDA_PIN = GPIO5
	SCL_PIN = GPIO6
)

// SPI pins
const (
	SPI_SCK_PIN = GPIO7
	SPI_SDI_PIN = GPIO9
	SPI_SDO_PIN = GPIO8
)

// Onboard LEDs
const (
	LED = GPIO21
)

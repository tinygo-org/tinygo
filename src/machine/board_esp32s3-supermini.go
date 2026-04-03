//go:build esp32s3_supermini

// This file contains the pin mappings for the ESP32S3 supermini board.
//
// - https://www.nologo.tech/product/esp32/esp32s3/esp32s3supermini/esp32S3SuperMini.html

package machine

// Digital Pins
const (
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
	IO11 = GPIO11
	IO12 = GPIO12
	IO13 = GPIO13
	IO14 = GPIO14
	IO15 = GPIO15
	IO16 = GPIO16
	IO17 = GPIO17
	IO18 = GPIO18
	IO21 = GPIO21
	IO33 = GPIO33
	IO34 = GPIO34
	IO35 = GPIO35
	IO36 = GPIO36
	IO37 = GPIO37
	IO38 = GPIO38
	IO39 = GPIO39
	IO40 = GPIO40
	IO41 = GPIO41
	IO42 = GPIO42
	IO43 = GPIO43
	IO44 = GPIO44
	IO45 = GPIO45
	IO46 = GPIO46
	IO47 = GPIO47
	IO48 = GPIO48
)

// Built-in LED
const LED = GPIO48

// Analog pins
const (
	A1  = GPIO1
	A2  = GPIO2
	A3  = GPIO3
	A4  = GPIO4
	A5  = GPIO5
	A6  = GPIO6
	A7  = GPIO7
	A8  = GPIO8
	A9  = GPIO9
	A10 = GPIO10
	A11 = GPIO11
	A12 = GPIO12
	A13 = GPIO13
	A14 = GPIO14
	A15 = GPIO15
	A16 = GPIO16
)

// I2C pins
const (
	SDA_PIN = GPIO8
	SCL_PIN = GPIO9
)

// SPI pins
const (
	SPI1_SCK_PIN  = GPIO4
	SPI1_MISO_PIN = GPIO5
	SPI1_MOSI_PIN = GPIO6
	SPI1_CS_PIN   = GPIO7

	SPI2_SCK_PIN  = NoPin
	SPI2_MOSI_PIN = NoPin
	SPI2_MISO_PIN = NoPin
	SPI2_CS_PIN   = NoPin
)

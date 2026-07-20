//go:build m5stamp_s3a

// This file contains the pin mappings for the M5Stack Stamp-S3A module.
//
// M5Stack Stamp-S3A is an embedded module based on the Espressif ESP32-S3FN8
// chip with 8MB flash, a programmable RGB LED, a user button, and 23 exposed
// GPIOs.
//
// - https://docs.m5stack.com/en/core/Stamp-S3A

package machine

// Digital pins
const (
	G0  = GPIO0
	G1  = GPIO1
	G2  = GPIO2
	G3  = GPIO3
	G4  = GPIO4
	G5  = GPIO5
	G6  = GPIO6
	G7  = GPIO7
	G8  = GPIO8
	G9  = GPIO9
	G10 = GPIO10
	G11 = GPIO11
	G12 = GPIO12
	G13 = GPIO13
	G14 = GPIO14
	G15 = GPIO15
	G39 = GPIO39
	G40 = GPIO40
	G41 = GPIO41
	G42 = GPIO42
	G43 = GPIO43
	G44 = GPIO44
	G46 = GPIO46
)

// Button
const (
	BUTTON = GPIO0
)

// UART pins
const (
	UART_RX_PIN = GPIO44
	UART_TX_PIN = GPIO43
)

// I2C pins
const (
	SDA_PIN = GPIO13
	SCL_PIN = GPIO14
)

// SPI pins
const (
	SPI1_SCK_PIN  = GPIO36
	SPI1_MISO_PIN = NoPin
	SPI1_MOSI_PIN = GPIO35
	SPI1_CS_PIN   = NoPin

	SPI2_SCK_PIN  = NoPin
	SPI2_MISO_PIN = NoPin
	SPI2_MOSI_PIN = NoPin
	SPI2_CS_PIN   = NoPin
)

// Display pins
const (
	DISPLAY_RST  = GPIO33
	DISPLAY_DC   = GPIO34
	DISPLAY_MOSI = GPIO35
	DISPLAY_SCK  = GPIO36
	DISPLAY_CS   = GPIO37
	DISPLAY_BL   = GPIO38
)

// Onboard LEDs
const (
	// WS2812 is the data pin for the onboard addressable RGB LED.
	WS2812 = GPIO21

	// WS2812_POWER enables power for the onboard addressable RGB LED.
	WS2812_POWER = GPIO38
)

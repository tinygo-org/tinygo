//go:build esp32c6

// This file contains the default pin mappings for the ESP32-C6-DevKitC target.

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
	IO11 = GPIO11
	IO12 = GPIO12
	IO13 = GPIO13
	IO14 = GPIO14
	IO15 = GPIO15
	IO16 = GPIO16
	IO17 = GPIO17
	IO18 = GPIO18
	IO19 = GPIO19
	IO20 = GPIO20
	IO21 = GPIO21
	IO22 = GPIO22
	IO23 = GPIO23
	IO24 = GPIO24
	IO25 = GPIO25
	IO26 = GPIO26
	IO27 = GPIO27
	IO28 = GPIO28
	IO29 = GPIO29
	IO30 = GPIO30
)

// Built-in WS2812 (NeoPixel) addressable RGB LED on the ESP32-C6-DevKitC.
// Use tinygo.org/x/drivers/ws2812 to control it.
const (
	LED      = WS2812
	WS2812   = GPIO8
	NEOPIXEL = GPIO8
)

// UART pins
const (
	UART_RX_PIN = GPIO17
	UART_TX_PIN = GPIO16
)

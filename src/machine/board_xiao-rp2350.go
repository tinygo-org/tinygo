//go:build xiao_rp2350

// This file contains the pin mappings for the Seeed XIAO RP2350 boards.
//
// XIAO RP2350 is a microcontroller using the Raspberry Pi RP2350 chip.
//
// - https://wiki.seeedstudio.com/XIAO-RP2350/
package machine

// Digital Pins
const (
	D0  Pin = GPIO26
	D1  Pin = GPIO27
	D2  Pin = GPIO28
	D3  Pin = GPIO5
	D4  Pin = GPIO6
	D5  Pin = GPIO7
	D6  Pin = GPIO0
	D7  Pin = GPIO1
	D8  Pin = GPIO2
	D9  Pin = GPIO4
	D10 Pin = GPIO3
	D11 Pin = GPIO21
	D12 Pin = GPIO20
	D13 Pin = GPIO17
	D14 Pin = GPIO16
	D15 Pin = GPIO11
	D16 Pin = GPIO12
	D17 Pin = GPIO10
)

// Analog pins
const (
	A0 Pin = D0
	A1 Pin = D1
	A2 Pin = D2
)

// Onboard LEDs
const (
	NEOPIXEL       = GPIO22
	WS2812         = GPIO22
	NEO_PWR        = GPIO23
	NEOPIXEL_POWER = GPIO23

	LED = GPIO25
)

// I2C pins
const (
	I2C0_SDA_PIN Pin = D14
	I2C0_SCL_PIN Pin = D13

	I2C1_SDA_PIN Pin = D4
	I2C1_SCL_PIN Pin = D5
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = D8
	SPI0_SDO_PIN Pin = D10
	SPI0_SDI_PIN Pin = D9

	SPI1_SCK_PIN Pin = D17
	SPI1_SDO_PIN Pin = D15
	SPI1_SDI_PIN Pin = D16
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "XIAO RP2350"
	usb_STRING_MANUFACTURER = "Seeed"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x000a
)

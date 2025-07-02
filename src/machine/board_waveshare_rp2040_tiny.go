//go:build waveshare_rp2040_tiny

// This file contains the pin mappings for the Waveshare RP2040-Tiny boards.
//
// Waveshare RP2040-Tiny is a microcontroller using the Raspberry Pi RP2040 chip.
//
// - https://www.waveshare.com/wiki/RP2040-Tiny
package machine

// Digital Pins
const (
	GP0  Pin = GPIO0
	GP1  Pin = GPIO1
	GP2  Pin = GPIO2
	GP3  Pin = GPIO3
	GP4  Pin = GPIO4
	GP5  Pin = GPIO5
	GP6  Pin = GPIO6
	GP7  Pin = GPIO7
	GP8  Pin = GPIO8
	GP9  Pin = GPIO9
	GP10 Pin = GPIO10
	GP11 Pin = GPIO11
	GP12 Pin = GPIO12
	GP13 Pin = GPIO13
	GP14 Pin = GPIO14
	GP15 Pin = GPIO15
	GP16 Pin = GPIO16
	GP17 Pin = NoPin
	GP18 Pin = NoPin
	GP19 Pin = NoPin
	GP20 Pin = NoPin
	GP21 Pin = NoPin
	GP22 Pin = NoPin
	GP23 Pin = NoPin
	GP24 Pin = GPIO24
	GP25 Pin = GPIO25
	GP26 Pin = GPIO26
	GP27 Pin = GPIO27
	GP28 Pin = GPIO28
	GP29 Pin = GPIO29
)

// Analog pins
const (
	A0 Pin = GP26
	A1 Pin = GP27
	A2 Pin = GP28
	A3 Pin = GP29
)

// Onboard LEDs
const (
	LED    = GP16
	WS2812 = GP16
)

// I2C pins
const (
	I2C0_SDA_PIN Pin = GP0
	I2C0_SCL_PIN Pin = GP1
	I2C1_SDA_PIN Pin = GP2
	I2C1_SCL_PIN Pin = GP3

	// default I2C0
	I2C_SDA_PIN Pin = I2C0_SDA_PIN
	I2C_SCL_PIN Pin = I2C0_SCL_PIN
)

// SPI pins
const (
	SPI0_RX_PIN  Pin = GP0
	SPI0_CSN_PIN Pin = GP1
	SPI0_SCK_PIN Pin = GP2
	SPI0_TX_PIN  Pin = GP3
	SPI0_SDO_PIN Pin = SPI0_TX_PIN
	SPI0_SDI_PIN Pin = SPI0_RX_PIN

	SPI1_RX_PIN  Pin = GP8
	SPI1_CSN_PIN Pin = GP9
	SPI1_SCK_PIN Pin = GP10
	SPI1_TX_PIN  Pin = GP11
	SPI1_SDO_PIN Pin = SPI1_TX_PIN
	SPI1_SDI_PIN Pin = SPI1_RX_PIN

	// default SPI0
	SPI_RX_PIN  Pin = SPI0_RX_PIN
	SPI_CSN_PIN Pin = SPI0_CSN_PIN
	SPI_SCK_PIN Pin = SPI0_SCK_PIN
	SPI_TX_PIN  Pin = SPI0_TX_PIN
	SPI_SDO_PIN Pin = SPI0_TX_PIN
	SPI_SDI_PIN Pin = SPI0_RX_PIN
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// UART pins
const (
	UART0_TX_PIN = GP0
	UART0_RX_PIN = GP1
	UART1_TX_PIN = GP8
	UART1_RX_PIN = GP9

	// default UART0
	UART_TX_PIN = UART0_TX_PIN
	UART_RX_PIN = UART0_RX_PIN
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "RP2040-Tiny"
	usb_STRING_MANUFACTURER = "Waveshare"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x0003
)

//go:build waveshare_rp2040_pizero

// This file contains the pin mappings for the Waveshare RP2040-PiZero board.
//
// Waveshare RP2040-PiZero uses the Raspberry Pi RP2040 chip.
//
// - https://www.waveshare.com/wiki/RP2040-PiZero
// - https://github.com/raspberrypi/pico-sdk/blob/master/src/boards/include/boards/waveshare_rp2040_pizero.h
package machine

// GPIO pins
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
	GP17 Pin = GPIO17
	GP18 Pin = GPIO18
	GP19 Pin = GPIO19
	GP20 Pin = GPIO20
	GP21 Pin = GPIO21
	GP22 Pin = GPIO22
	GP23 Pin = GPIO23
	GP24 Pin = GPIO24
	GP25 Pin = GPIO25
	GP26 Pin = GPIO26
	GP27 Pin = GPIO27
	GP28 Pin = GPIO28
	GP29 Pin = GPIO29

	// The Pico SDK board header defines no default LED pin.
	LED Pin = NoPin

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

// Analog pins
const (
	A0 Pin = GP26
	A1 Pin = GP27
	A2 Pin = GP28
	A3 Pin = GP29
)

// I2C default pins
const (
	I2C0_SDA_PIN Pin = NoPin
	I2C0_SCL_PIN Pin = NoPin

	I2C1_SDA_PIN Pin = GP6
	I2C1_SCL_PIN Pin = GP7
)

// SPI default pins
const (
	SPI0_RX_PIN  Pin = GP16
	SPI0_CSN_PIN Pin = GP17
	SPI0_SCK_PIN Pin = GP18
	SPI0_TX_PIN  Pin = GP19
	SPI0_SDO_PIN Pin = SPI0_TX_PIN
	SPI0_SDI_PIN Pin = SPI0_RX_PIN

	SPI1_SCK_PIN Pin = GP10
	SPI1_SDO_PIN Pin = GP11
	SPI1_SDI_PIN Pin = GP12
)

// UART pins
const (
	UART0_TX_PIN = GP0
	UART0_RX_PIN = GP1
	UART1_TX_PIN = GP8
	UART1_RX_PIN = GP9
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0

// USB identifiers
const (
	usb_STRING_PRODUCT      = "RP2040-PiZero"
	usb_STRING_MANUFACTURER = "Waveshare"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x0003
)

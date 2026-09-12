//go:build waveshare_rp2350_pizero

// This file contains the pin mappings for the Waveshare RP2350-PiZero board.
//
// Waveshare RP2350-PiZero uses the Raspberry Pi RP2350B chip.
//
// - https://www.waveshare.com/wiki/RP2350-PiZero
// - https://github.com/raspberrypi/pico-sdk/blob/master/src/boards/include/boards/waveshare_rp2350_pizero.h
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
	GP30 Pin = GPIO30
	GP31 Pin = GPIO31
	GP32 Pin = GPIO32
	GP33 Pin = GPIO33
	GP34 Pin = GPIO34
	GP35 Pin = GPIO35
	GP36 Pin = GPIO36
	GP37 Pin = GPIO37
	GP38 Pin = GPIO38
	GP39 Pin = GPIO39
	GP40 Pin = GPIO40
	GP41 Pin = GPIO41
	GP42 Pin = GPIO42
	GP43 Pin = GPIO43
	GP44 Pin = GPIO44
	GP45 Pin = GPIO45
	GP46 Pin = GPIO46
	GP47 Pin = GPIO47

	// The Pico SDK board header defines no default LED pin.
	LED Pin = NoPin

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

// I2C default pins
const (
	I2C0_SDA_PIN = NoPin
	I2C0_SCL_PIN = NoPin

	I2C1_SDA_PIN = GP2
	I2C1_SCL_PIN = GP3
)

// SPI default pins
const (
	SPI0_SCK_PIN = NoPin
	SPI0_SDO_PIN = NoPin
	SPI0_SDI_PIN = NoPin

	SPI1_SCK_PIN = GPIO10
	SPI1_SDO_PIN = GPIO11 // Tx
	SPI1_SDI_PIN = GPIO12 // Rx
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART1_TX_PIN = GPIO4
	UART1_RX_PIN = GPIO5
	UART_TX_PIN  = UART1_TX_PIN
	UART_RX_PIN  = UART1_RX_PIN
)

var DefaultUART = UART1

// USB identifiers
const (
	usb_STRING_PRODUCT      = "RP2350-PiZero"
	usb_STRING_MANUFACTURER = "Waveshare"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x000f
)

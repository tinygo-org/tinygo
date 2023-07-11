//go:build tufty2040

// This contains the pin mappings for the Badger 2040 Connect board.
//
// For more information, see: https://shop.pimoroni.com/products/tufty-2040
// Also
// - Tufty 2040 schematic: https://cdn.shopify.com/s/files/1/0174/1800/files/tufty_schematic.pdf?v=1655385675
package machine

import (
	"device/rp"
	"runtime/interrupt"
)

const (
	LED Pin = GPIO25

	BUTTON_A    Pin = GPIO7
	BUTTON_B    Pin = GPIO8
	BUTTON_C    Pin = GPIO9
	BUTTON_UP   Pin = GPIO22
	BUTTON_DOWN Pin = GPIO6
	BUTTON_USER Pin = GPIO23

	LCD_BACKLIGHT Pin = GPIO2
	LCD_CS        Pin = GPIO10
	LCD_DC        Pin = GPIO11
	LCD_WR        Pin = GPIO12
	LCD_RD        Pin = GPIO13
	LCD_DB0       Pin = GPIO14
	LCD_DB1       Pin = GPIO15
	LCD_DB2       Pin = GPIO16
	LCD_DB3       Pin = GPIO17
	LCD_DB4       Pin = GPIO18
	LCD_DB5       Pin = GPIO19
	LCD_DB6       Pin = GPIO20
	LCD_DB7       Pin = GPIO21

	VBUS_DETECT  Pin = GPIO24
	BATTERY      Pin = GPIO29
	USER_LED     Pin = GPIO25
	LIGHT_SENSE  Pin = GPIO26
	SENSOR_POWER Pin = GPIO27
)

// I2C pins
const (
	I2C0_SDA_PIN Pin = GPIO4
	I2C0_SCL_PIN Pin = GPIO5

	I2C1_SDA_PIN Pin = NoPin
	I2C1_SCL_PIN Pin = NoPin
)

// SPI pins.
const (
	SPI0_SCK_PIN Pin = NoPin
	SPI0_SDO_PIN Pin = NoPin
	SPI0_SDI_PIN Pin = NoPin

	SPI1_SCK_PIN Pin = NoPin
	SPI1_SDO_PIN Pin = NoPin
	SPI1_SDI_PIN Pin = NoPin
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Tufty 2040"
	usb_STRING_MANUFACTURER = "Pimoroni"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x1002
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

// UART on the RP2040
var (
	UART0  = &_UART0
	_UART0 = UART{
		Buffer: NewRingBuffer(),
		Bus:    rp.UART0,
	}
)

var DefaultUART = UART0

func init() {
	UART0.Interrupt = interrupt.New(rp.IRQ_UART0_IRQ, _UART0.handleInterrupt)
}

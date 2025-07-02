//go:build badger2040_w

// This contains the pin mappings for the Badger 2040 W board.
//
// For more information, see: https://shop.pimoroni.com/products/badger-2040-w
// Also
// - Badger 2040 W schematic: https://cdn.shopify.com/s/files/1/0174/1800/files/badger_w_schematic.pdf?v=1675859004
package machine

const (
	LED Pin = GPIO22

	BUTTON_A    Pin = GPIO12
	BUTTON_B    Pin = GPIO13
	BUTTON_C    Pin = GPIO14
	BUTTON_UP   Pin = GPIO15
	BUTTON_DOWN Pin = GPIO11
	BUTTON_USER Pin = NoPin // Not available on Badger 2040 W

	EPD_BUSY_PIN  Pin = GPIO26
	EPD_RESET_PIN Pin = GPIO21
	EPD_DC_PIN    Pin = GPIO20
	EPD_CS_PIN    Pin = GPIO17
	EPD_SCK_PIN   Pin = GPIO18
	EPD_SDO_PIN   Pin = GPIO19

	VBUS_DETECT Pin = GPIO24
	VREF_POWER  Pin = GPIO27
	VREF_1V24   Pin = GPIO28
	VBAT_SENSE  Pin = GPIO29
	ENABLE_3V3  Pin = GPIO10

	BATTERY   = VBAT_SENSE
	RTC_ALARM = GPIO8
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
	SPI0_SCK_PIN Pin = GPIO18
	SPI0_SDO_PIN Pin = GPIO19
	SPI0_SDI_PIN Pin = GPIO16

	SPI1_SCK_PIN Pin = NoPin
	SPI1_SDO_PIN Pin = NoPin
	SPI1_SDI_PIN Pin = NoPin
)

// QSPI pins¿?
const (
/*
	TODO

SPI0_SD0_PIN Pin = QSPI_SD0
SPI0_SD1_PIN Pin = QSPI_SD1
SPI0_SD2_PIN Pin = QSPI_SD2
SPI0_SD3_PIN Pin = QSPI_SD3
SPI0_SCK_PIN Pin = QSPI_SCLKGPIO6
SPI0_CS_PIN  Pin = QSPI_CS
*/
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Badger 2040 W"
	usb_STRING_MANUFACTURER = "Pimoroni"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x0003
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0

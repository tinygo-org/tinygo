//go:build thumby

// This contains the pin mappings for the Thumby.
//
// https://thumby.us/
package machine

const (
	THUMBY_SCK_PIN = I2C1_SDA_PIN
	THUMBY_SDA_PIN = I2C1_SCL_PIN

	THUMBY_CS_PIN    = GPIO16
	THUMBY_DC_PIN    = GPIO17
	THUMBY_RESET_PIN = GPIO20

	THUMBY_LINK_TX_PIN = UART0_TX_PIN
	THUMBY_LINK_RX_PIN = UART0_RX_PIN
	THUMBY_LINK_PU_PIN = GPIO2

	THUMBY_BTN_LDPAD_PIN = GPIO3
	THUMBY_BTN_RDPAD_PIN = GPIO5
	THUMBY_BTN_UDPAD_PIN = GPIO4
	THUMBY_BTN_DDPAD_PIN = GPIO6
	THUMBY_BTN_B_PIN     = GPIO24
	THUMBY_BTN_A_PIN     = GPIO27

	THUMBY_AUDIO_PIN = GPIO28

	THUMBY_SCREEN_RESET_PIN = GPIO20
)

// I2C pins
const (
	I2C0_SDA_PIN Pin = NoPin
	I2C0_SCL_PIN Pin = NoPin

	I2C1_SDA_PIN Pin = GPIO18
	I2C1_SCL_PIN Pin = GPIO19
)

// SPI pins
const (
	SPI0_SCK_PIN = GPIO18
	SPI0_SDO_PIN = GPIO19
	SPI0_SDI_PIN = GPIO16

	SPI1_SCK_PIN = NoPin
	SPI1_SDO_PIN = NoPin
	SPI1_SDI_PIN = NoPin
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Thumby"
	usb_STRING_MANUFACTURER = "TinyCircuits"
)

var (
	usb_VID uint16 = 0x2E8A
	usb_PID uint16 = 0x0005
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0

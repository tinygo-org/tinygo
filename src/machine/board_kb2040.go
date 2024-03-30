//go:build kb2040

package machine

// Onboard crystal oscillator frequency, in MHz.
const xoscFreq = 12 // MHz

// GPIO Pins
const (
	D0  = GPIO0
	D1  = GPIO1
	D2  = GPIO2
	D3  = GPIO3
	D4  = GPIO4
	D5  = GPIO5
	D6  = GPIO6
	D7  = GPIO7
	D8  = GPIO8
	D9  = GPIO9
	D10 = GPIO10
)

// Analog pins
const (
	A0 = GPIO26
	A1 = GPIO27
	A2 = GPIO28
	A3 = GPIO29
)

// Note: there is no user-controllable LED on the KB2040 board
// const LED = notConnected

// I2C Pins.
const (
	I2C0_SDA_PIN = GPIO12
	I2C0_SCL_PIN = GPIO13

	I2C1_SDA_PIN = GPIO2
	I2C1_SCL_PIN = GPIO3

	SDA_PIN = I2C0_SDA_PIN
	SCL_PIN = I2C0_SCL_PIN
)

// SPI default pins
const (
	// Default Serial Clock Bus 0 for SPI communications
	SPI0_SCK_PIN = GPIO18
	// Default Serial Out Bus 0 for SPI communications
	SPI0_SDO_PIN = GPIO19 // Tx
	// Default Serial In Bus 0 for SPI communications
	SPI0_SDI_PIN = GPIO20 // Rx

	// Default Serial Clock Bus 1 for SPI communications
	SPI1_SCK_PIN = GPIO26
	// Default Serial Out Bus 1 for SPI communications
	SPI1_SDO_PIN = GPIO27 // Tx
	// Default Serial In Bus 1 for SPI communications
	SPI1_SDI_PIN = GPIO28 // Rx
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART1_TX_PIN = GPIO8
	UART1_RX_PIN = GPIO9
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0

// USB identifiers
const (
	usb_STRING_PRODUCT      = "KB2040"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x8106
)

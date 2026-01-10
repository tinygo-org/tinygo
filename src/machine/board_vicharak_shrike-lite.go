//go:build vicharak_shrike_lite

// Pin mappings for Vicharak Shrike-Lite.
//
// Reference: https://vicharak-in.github.io/shrike/shrike_pinouts.html

package machine

// Digital
const (
	IO0  Pin = GPIO0
	IO1  Pin = GPIO1
	IO2  Pin = GPIO2
	IO3  Pin = GPIO3
	IO4  Pin = GPIO4
	IO5  Pin = GPIO5
	IO6  Pin = GPIO6
	IO7  Pin = GPIO7
	IO8  Pin = GPIO8
	IO9  Pin = GPIO9
	IO10 Pin = GPIO10
	IO11 Pin = GPIO11
	IO12 Pin = GPIO12
	IO13 Pin = GPIO13
	IO14 Pin = GPIO14
	IO15 Pin = GPIO15
	IO16 Pin = GPIO16
	IO17 Pin = GPIO17
	IO18 Pin = GPIO18
	IO19 Pin = GPIO19
	IO20 Pin = GPIO20
	IO21 Pin = GPIO21
	IO22 Pin = GPIO22
	IO23 Pin = GPIO23
	IO24 Pin = GPIO24
	IO25 Pin = GPIO25
	IO26 Pin = GPIO26
	IO27 Pin = GPIO27
	IO28 Pin = GPIO28
	IO29 Pin = GPIO29
)

// FPGA Pins
const (
	FPGA_EN  Pin = IO13
	FPGA_PWR Pin = IO12
	// SPI_SCLK
	F3 Pin = IO2
	// SPI_SS
	F4 Pin = IO1
	// SPI_SI (MOSI)
	F5 Pin = IO3
	// SPI_SO (MISO) / CONFIG
	F6  Pin = IO0
	F18 Pin = IO14
	F17 Pin = IO15
)

// Analog pins
const (
	A0 Pin = IO26
	A1 Pin = IO27
	A2 Pin = IO28
	A3 Pin = IO29
)

// LED
const (
	LED = IO4
)

// I2C pins
const (
	I2C0_SDA_PIN Pin = IO24
	I2C0_SCL_PIN Pin = IO25

	I2C1_SDA_PIN Pin = IO6
	I2C1_SCL_PIN Pin = IO7
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = IO18
	SPI0_SDO_PIN Pin = IO19
	SPI0_SDI_PIN Pin = IO20

	SPI1_SCK_PIN Pin = IO10
	SPI1_SDO_PIN Pin = IO11
	SPI1_SDI_PIN Pin = IO8
)

// Onboard crystal oscillator frequency, in MHz.
const (
	xoscFreq = 12 // MHz
)

// UART pins
const (
	UART0_TX_PIN = IO28
	UART0_RX_PIN = IO29
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
	UART1_TX_PIN = IO24
	UART1_RX_PIN = IO25
)

var DefaultUART = UART0

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Shrike-Lite"
	usb_STRING_MANUFACTURER = "Vicharak"
)

var (
	usb_VID uint16 = 0x2e8a
	usb_PID uint16 = 0x0003
)

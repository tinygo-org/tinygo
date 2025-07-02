//go:build tiny2350

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
	GP12 Pin = GPIO12
	GP13 Pin = GPIO13
	GP18 Pin = GPIO18
	GP19 Pin = GPIO19
	GP20 Pin = GPIO20
	GP26 Pin = GPIO26
	GP27 Pin = GPIO27
	GP28 Pin = GPIO28
	GP29 Pin = GPIO29

	// Onboard LED
	LED_RED   Pin = GPIO18
	LED_GREEN Pin = GPIO19
	LED_BLUE  Pin = GPIO20
	LED           = LED_RED

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

// I2C Default pins on Tiny2350.
const (
	I2C0_SDA_PIN = GP12
	I2C0_SCL_PIN = GP13

	I2C1_SDA_PIN = GP2
	I2C1_SCL_PIN = GP3
)

// SPI default pins
const (
	// Default Serial Clock Bus 0 for SPI communications
	SPI0_SCK_PIN = GPIO6
	// Default Serial Out Bus 0 for SPI communications
	SPI0_SDO_PIN = GPIO7 // Tx
	// Default Serial In Bus 0 for SPI communications
	SPI0_SDI_PIN = GPIO4 // Rx

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
	UART1_TX_PIN = GPIO4
	UART1_RX_PIN = GPIO5
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0

// USB identifiers
const (
	usb_STRING_PRODUCT      = "Tiny2350"
	usb_STRING_MANUFACTURER = "Pimoroni"
)

var (
	usb_VID uint16 = 0x2E8A
	usb_PID uint16 = 0x000F
)

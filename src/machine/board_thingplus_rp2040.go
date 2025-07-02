//go:build thingplus_rp2040

package machine

// Onboard crystal oscillator frequency, in MHz.
const xoscFreq = 12 // MHz

// GPIO Pins
const (
	GP0 Pin = GPIO0 // TX
	GP1 Pin = GPIO1 // RX
	GP2 Pin = GPIO2 // SCK
	GP3 Pin = GPIO3 // COPI
	GP4 Pin = GPIO4 // CIPO

	GP6  Pin = GPIO6  // SDA
	GP7  Pin = GPIO7  // SCL (connected to GPIO23 as well)
	GP8  Pin = GPIO8  // WS2812 RGB LED
	GP9  Pin = GPIO9  // muSDcard DATA3 / CS
	GP10 Pin = GPIO10 // muSDcard DATA2
	GP11 Pin = GPIO11 // muSDcard DATA1
	GP12 Pin = GPIO12 // muSDcard DATA0 / CIPO

	GP14 Pin = GPIO14 // muSDcard CLK /SCLK
	GP15 Pin = GPIO15 // muSDcard CMD / COPI
	GP16 Pin = GPIO16 // 16
	GP17 Pin = GPIO17 // 17
	GP18 Pin = GPIO18 // 18
	GP19 Pin = GPIO19 // 19
	GP20 Pin = GPIO20 // 20
	GP21 Pin = GPIO21 // 21
	GP22 Pin = GPIO22 // 22
	GP23 Pin = GPIO23 // Connected to GPIO7
	GP25 Pin = GPIO25 // Status blue LED
	GP26 Pin = GPIO26 // ADC0
	GP27 Pin = GPIO27 // ADC1
	GP28 Pin = GPIO28 // ADC2
	GP29 Pin = GPIO29 // ADC3
)

// Analog pins
const (
	A0 = GPIO26
	A1 = GPIO27
	A2 = GPIO28
	A3 = GPIO29
)

// Onboard LEDs
const (
	LED    = GPIO25
	WS2812 = GPIO8
)

// I2C Pins.
const (
	I2C0_SCL_PIN = GPIO6 // N/A
	I2C0_SDA_PIN = GPIO7 // N/A

	I2C1_SDA_PIN = GPIO6
	I2C1_SCL_PIN = GPIO7

	SDA_PIN = I2C1_SDA_PIN
	SCL_PIN = I2C1_SCL_PIN
)

// SPI default pins
const (
	// Default Serial Clock Bus 0 for SPI communications
	SPI0_SCK_PIN = GPIO2
	// Default Serial Out Bus 0 for SPI communications
	SPI0_SDO_PIN = GPIO3 // Tx
	// Default Serial In Bus 0 for SPI communications
	SPI0_SDI_PIN = GPIO4 // Rx

	// Default Serial Clock Bus 1 for SPI communications to muSDcard
	SPI1_SCK_PIN = GPIO14
	// Default Serial Out Bus 1 for SPI communications to muSDcard
	SPI1_SDO_PIN = GPIO15 // Tx
	// Default Serial In Bus 1 for SPI communications to muSDcard
	SPI1_SDI_PIN = GPIO12 // Rx
)

// UART pins
const (
	UART0_TX_PIN = GPIO0
	UART0_RX_PIN = GPIO1
	UART_TX_PIN  = UART0_TX_PIN
	UART_RX_PIN  = UART0_RX_PIN
)

var DefaultUART = UART0

// USB identifiers
const (
	usb_STRING_PRODUCT      = "Thing Plus RP2040"
	usb_STRING_MANUFACTURER = "SparkFun"
)

var (
	usb_VID uint16 = 0x1B4F
	usb_PID uint16 = 0x0026
)

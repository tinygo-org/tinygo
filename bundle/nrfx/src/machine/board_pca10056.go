//go:build pca10056

package machine

const HasLowFrequencyCrystal = true

// LEDs on the pca10056
const (
	LED1 Pin = 13
	LED2 Pin = 14
	LED3 Pin = 15
	LED4 Pin = 16
	LED  Pin = LED1
)

// Buttons on the pca10056
const (
	BUTTON1 Pin = 11
	BUTTON2 Pin = 12
	BUTTON3 Pin = 24
	BUTTON4 Pin = 25
	BUTTON  Pin = BUTTON1
)

var DefaultUART = UART0

// UART pins
const (
	UART_TX_PIN Pin = 6
	UART_RX_PIN Pin = 8
)

// ADC pins
const (
	ADC0 Pin = 3
	ADC1 Pin = 4
	ADC2 Pin = 28
	ADC3 Pin = 29
	ADC4 Pin = 30
	ADC5 Pin = 31
)

// I2C pins
const (
	SDA_PIN Pin = 26 // P0.26
	SCL_PIN Pin = 27 // P0.27
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = 47 // P1.15
	SPI0_SDO_PIN Pin = 45 // P1.13
	SPI0_SDI_PIN Pin = 46 // P1.14
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Nordic nRF52840DK (PCA10056)"
	usb_STRING_MANUFACTURER = "Nordic Semiconductor"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x8029
)

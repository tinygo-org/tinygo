//go:build pca10059
// +build pca10059

package machine

// The PCA10040 has a low-frequency (32kHz) crystal oscillator on board.
const HasLowFrequencyCrystal = true

// LEDs on the PCA10059 (nRF52840 dongle)
const (
	LED1 Pin = 6
	LED2 Pin = 8
	LED3 Pin = (1 << 5) | 9
	LED4 Pin = 12
	LED  Pin = LED1
)

// Buttons on the PCA10059 (nRF52840 dongle)
const (
	BUTTON1 Pin = (1 << 5) | 6
	BUTTON  Pin = BUTTON1
)

// ADC pins
const (
	ADC1 Pin = 2
	ADC2 Pin = 4
	ADC3 Pin = 29
	ADC4 Pin = 31
)

// UART pins
const (
	UART_TX_PIN Pin = NoPin
	UART_RX_PIN Pin = NoPin
)

// I2C pins (unused)
const (
	SDA_PIN = NoPin
	SCL_PIN = NoPin
)

// SPI pins (unused)
const (
	SPI0_SCK_PIN = NoPin
	SPI0_SDO_PIN = NoPin
	SPI0_SDI_PIN = NoPin
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "nRF52840 Dongle"
	usb_STRING_MANUFACTURER = "Nordic Semiconductor ASA"
)

var (
	usb_VID uint16 = 0x1915
	usb_PID uint16 = 0xCAFE
)

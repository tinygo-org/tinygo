//go:build nrf52840_mdk_usb_dongle

package machine

const HasLowFrequencyCrystal = true

// LEDs on the nrf52840-mdk-usb-dongle
const (
	LED       Pin = LED_GREEN
	LED_GREEN Pin = 22
	LED_RED   Pin = 23
	LED_BLUE  Pin = 24
)

// RESET/USR button, depending on value of PSELRESET UICR register
const (
	BUTTON Pin = 18
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
	usb_STRING_PRODUCT      = "Makerdiary nRF52840 MDK USB Dongle"
	usb_STRING_MANUFACTURER = "Nordic Semiconductor ASA"
)

var (
	usb_VID uint16 = 0x1915
	usb_PID uint16 = 0xCAFE
)

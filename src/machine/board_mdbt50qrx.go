//go:build mdbt50qrx
// +build mdbt50qrx

package machine

const HasLowFrequencyCrystal = false

// GPIO Pins
const (
	D0 = P1_13 // LED1
	D1 = P1_11 // LED2 (not populated by default)
	D2 = P0_15 // Button
)

const (
	LED = D0
)

// MDBT50Q-RX dongle does not have pins broken out for the peripherals below,
// however the machine_nrf*.go implementations of I2C/SPI/etc expect the pin
// constants to be defined, so we are defining them all as 0
const (
	UART_TX_PIN  = 0
	UART_RX_PIN  = 0
	SDA_PIN      = 0
	SCL_PIN      = 0
	SPI0_SCK_PIN = 0
	SPI0_SDO_PIN = 0
	SPI0_SDI_PIN = 0
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Raytac MDBT50Q - RX"
	usb_STRING_MANUFACTURER = "Raytac Corporation"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x810B
)

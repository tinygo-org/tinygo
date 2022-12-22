//go:build sam && atsamd21 && qtpy

package machine

// used to reset into bootloader
const resetMagicValue = 0xf01669ef

// GPIO Pins
const (
	D0  = PA02 // PWM available
	D1  = PA03
	D2  = PA04 // PWM available
	D3  = PA05 // PWM available
	D4  = PA16 // PWM available
	D5  = PA17 // PWM available
	D6  = PA06
	D7  = PA07
	D8  = PA11
	D9  = PA09
	D10 = PA10
	D11 = PA18
	D12 = PA15
	D13 = PA27
	D14 = PA23
	D15 = PA19
	D16 = PA22
	D17 = PA08
)

// Analog pins
const (
	A0 = D1
	A1 = D1
	A2 = D2
	A3 = D3
	A4 = D4
)

const (
	NEOPIXELS       = D11
	WS2812          = D11
	NEOPIXELS_POWER = D12
)

// USBCDC pins
const (
	USBCDC_DM_PIN = PA24
	USBCDC_DP_PIN = PA25
)

// UART1 pins
const (
	UART_TX_PIN = D6
	UART_RX_PIN = D7
)

// UART1 on the QT Py M0.
var UART1 = &sercomUSART0

// SPI pins
const (
	SPI0_SCK_PIN = D8
	SPI0_SDO_PIN = D10
	SPI0_SDI_PIN = D9
)

// SPI on the QT Py M0.
var SPI0 = sercomSPIM0

// I2C pins
const (
	SDA_PIN = D4 // SDA
	SCL_PIN = D5 // SCL
)

// I2C on the QT Py M0.
var (
	I2C0 = sercomI2CM1
)

// I2S pins
const (
	I2S_SCK_PIN = PA10
	I2S_SD_PIN  = PA08
	I2S_WS_PIN  = NoPin // TODO: figure out what this is on QT Py M0.
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit QTPy M0"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x80CB
)

var (
	DefaultUART = UART1
)

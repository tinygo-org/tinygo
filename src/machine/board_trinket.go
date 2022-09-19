//go:build sam && atsamd21 && trinket_m0
// +build sam,atsamd21,trinket_m0

package machine

// used to reset into bootloader
const resetMagicValue = 0xf01669ef

// GPIO Pins
const (
	D0  = PA08 // PWM available
	D1  = PA02
	D2  = PA09 // PWM available
	D3  = PA07 // PWM available / UART0 RX
	D4  = PA06 // PWM available / UART0 TX
	D13 = PA10 // LED
)

// Analog pins
const (
	A0 = D1
	A1 = D2
	A2 = D0
	A3 = D3
	A4 = D4
)

const (
	LED = D13
)

// USBCDC pins
const (
	USBCDC_DM_PIN = PA24
	USBCDC_DP_PIN = PA25
)

// UART1 pins
const (
	UART_TX_PIN = D4
	UART_RX_PIN = D3
)

// UART1 on the Trinket M0.
var UART1 = &sercomUSART0

// SPI pins
const (
	SPI0_SCK_PIN = D3
	SPI0_SDO_PIN = D4
	SPI0_SDI_PIN = D2
)

// SPI on the Trinket M0.
var SPI0 = sercomSPIM0

// I2C pins
const (
	SDA_PIN = D0 // SDA
	SCL_PIN = D2 // SCL
)

// I2C on the Trinket M0.
var (
	I2C0 = sercomI2CM2
)

// I2S pins
const (
	I2S_SCK_PIN = PA10
	I2S_SD_PIN  = PA08
	I2S_WS_PIN  = NoPin // TODO: figure out what this is on Trinket M0.
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit Trinket M0"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x801E
)

var (
	DefaultUART = UART1
)

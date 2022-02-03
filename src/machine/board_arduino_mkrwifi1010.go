//go:build arduino_mkrwifi1010
// +build arduino_mkrwifi1010

// This contains the pin mappings for the Arduino MKR WiFi 1010 board.
//
// For more information, see: https://store.arduino.cc/usa/mkr-wifi-1010
//
package machine

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0x07738135

// GPIO Pins
const (
	RX0 Pin = PB23 // UART1 RX
	TX1 Pin = PB22 // UART1 TX

	D0 Pin = PA22 // PWM available
	D1 Pin = PA23 // PWM available
	D2 Pin = PA10 // PWM available
	D3 Pin = PA11 // PWM available
	D4 Pin = PB10 // PWM available
	D5 Pin = PB11 // PWM available

	D6  Pin = PA20 // PWM available
	D7  Pin = PA21 // PWM available
	D8  Pin = PA16 // PWM available
	D9  Pin = PA17
	D10 Pin = PA19 // PWM available
	D11 Pin = PA08 // SDA
	D12 Pin = PA09 // PWM available, SCL
	D13 Pin = PB23 // RX
	D14 Pin = PB22 // TX
)

// Analog pins
const (
	A0 Pin = PA02 // ADC0/AIN[0]
	A1 Pin = PB02 // AIN[10]
	A2 Pin = PB03 // AIN[11]
	A3 Pin = PA04 // AIN[04]
	A4 Pin = PA05 // AIN[05]
	A5 Pin = PA06 // AIN[06]
	A6 Pin = PA07 // AIN[07]
)

const (
	LED = D6
)

// USBCDC pins
const (
	USBCDC_DM_PIN Pin = PA24
	USBCDC_DP_PIN Pin = PA25
)

// UART1 pins
const (
	UART_TX_PIN Pin = PB22
	UART_RX_PIN Pin = PB23
)

// I2C pins
const (
	SDA_PIN Pin = D11 // SDA
	SCL_PIN Pin = D12 // SCL
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = D9  // SCK: S1
	SPI0_SDO_PIN Pin = D8  // SDO: S1
	SPI0_SDI_PIN Pin = D10 // SDI: S1
)

// I2S pins
const (
	I2S_SCK_PIN Pin = PA10
	I2S_SD_PIN  Pin = PA07
	I2S_WS_PIN      = NoPin // TODO: figure out what this is on Arduino MKR WiFi 1010.
)

// NINA-W102 Pins
const (
	NINA_SDO    Pin = PA12
	NINA_SDI    Pin = PA13
	NINA_CS     Pin = PA14
	NINA_SCK    Pin = PA15
	NINA_GPIO0  Pin = PA27
	NINA_RESETN Pin = PB08
	NINA_ACK    Pin = PA28
	NINA_TX     Pin = PA22
	NINA_RX     Pin = PA23
)

// UART on the Arduino MKR WiFi 1010.
var UART1 = &sercomUSART5

// I2C on the Arduino MKR WiFi 1010.
var (
	I2C0 = sercomI2CM2
)

// SPI on the Arduino MKR WiFi 1010.
var (
	SPI0 = sercomSPIM1

	SPI1     = sercomSPIM4
	NINA_SPI = SPI1
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Arduino MKR WiFi 1010"
	usb_STRING_MANUFACTURER = "Arduino"
)

var (
	usb_VID uint16 = 0x2341
	usb_PID uint16 = 0x8054
)

var (
	DefaultUART = UART1
)

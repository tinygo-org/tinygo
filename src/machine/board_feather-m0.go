//go:build sam && atsamd21 && feather_m0
// +build sam,atsamd21,feather_m0

package machine

// used to reset into bootloader
const resetMagicValue = 0xf01669ef

// GPIO Pins
const (
	D0  = PA11  // UART0 RX
	D1  = PA10  // UART0 TX
	D2  = NoPin // does not seem to exist
	D3  = PA09
	D4  = PA08
	D5  = PA15  // PWM available
	D6  = PA20  // PWM available
	D7  = NoPin // does not seem to exist
	D8  = PA06
	D9  = PA07 // PWM available
	D10 = PA18 // can be used for PWM or UART1 TX
	D11 = PA16 // can be used for PWM or UART1 RX
	D12 = PA19 // PWM available
	D13 = PA17 // PWM available
)

// Analog pins
const (
	A0 = PA02 // ADC/AIN[0]
	A1 = PB08 // ADC/AIN[2]
	A2 = PB09 // ADC/AIN[3]
	A3 = PA04 // ADC/AIN[4]
	A4 = PA05 // ADC/AIN[5]
	A5 = PB02 // ADC/AIN[10]
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
	UART_TX_PIN = D10
	UART_RX_PIN = D11
)

// UART1 on the Feather M0.
var UART1 = &sercomUSART1

// I2C pins
const (
	SDA_PIN = PA22 // SDA: SERCOM3/PAD[0]
	SCL_PIN = PA23 // SCL: SERCOM3/PAD[1]
)

// I2C on the Feather M0.
var (
	I2C0 = sercomI2CM3
)

// SPI pins
const (
	SPI0_SCK_PIN = PB11 // SCK: SERCOM4/PAD[3]
	SPI0_SDO_PIN = PB10 // SDO: SERCOM4/PAD[2]
	SPI0_SDI_PIN = PA12 // SDI: SERCOM4/PAD[0]
)

// SPI on the Feather M0.
var SPI0 = sercomSPIM4

// I2S pins
const (
	I2S_SCK_PIN = PA10
	I2S_SD_PIN  = PA08
	I2S_WS_PIN  = NoPin // TODO: figure out what this is on Feather M0.
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit Feather M0 Express"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x801B
)

var (
	DefaultUART = UART1
)

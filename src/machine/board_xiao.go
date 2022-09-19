//go:build sam && atsamd21 && xiao
// +build sam,atsamd21,xiao

package machine

// used to reset into bootloader
const resetMagicValue = 0xf01669ef

// GPIO Pins
const (
	D0  = PA02 // can be used for PWM or DAC
	D1  = PA04 // PWM available
	D2  = PA10 // PWM available
	D3  = PA11 // PWM available
	D4  = PA08 // can be used for PWM or I2C SDA
	D5  = PA09 // can be used for PWM or I2C SCL
	D6  = PB08 // can be used for PWM or UART1 TX
	D7  = PB09 // can be used for PWM or UART1 RX
	D8  = PA07 // can be used for PWM or SPI SCK
	D9  = PA05 // can be used for PWM or SPI SDI
	D10 = PA06 // can be used for PWM or SPI SDO
)

// Analog pins
const (
	A0  = PA02 // ADC/AIN[0]
	A1  = PA04 // ADC/AIN[4]
	A2  = PA10 // ADC/AIN[18]
	A3  = PA11 // ADC/AIN[19]
	A4  = PA08 // ADC/AIN[16]
	A5  = PA09 // ADC/AIN[17]
	A6  = PB08 // ADC/AIN[2]
	A7  = PB09 // ADC/AIN[3]
	A8  = PA07 // ADC/AIN[7]
	A9  = PA05 // ADC/AIN[6]
	A10 = PA06 // ADC/AIN[5]
)

const (
	LED     = PA17
	LED_RXL = PA18
	LED_TXL = PA19
	LED2    = LED_RXL
	LED3    = LED_TXL
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

// UART1 on the Xiao
var UART1 = &sercomUSART4

// I2C pins
const (
	SDA_PIN = PA08 // SDA: SERCOM2/PAD[0]
	SCL_PIN = PA09 // SCL: SERCOM2/PAD[1]
)

// I2C on the Xiao
var (
	I2C0 = sercomI2CM2
)

// SPI pins
const (
	SPI0_SCK_PIN = PA07 // SCK: SERCOM0/PAD[3]
	SPI0_SDO_PIN = PA06 // SDO: SERCOM0/PAD[2]
	SPI0_SDI_PIN = PA05 // SDI: SERCOM0/PAD[1]
)

// SPI on the Xiao
var SPI0 = sercomSPIM0

// I2S pins
const (
	I2S_SCK_PIN = PA10
	I2S_SD_PIN  = PA08
	I2S_WS_PIN  = NoPin // TODO: figure out what this is on Xiao
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Seeed XIAO M0"
	usb_STRING_MANUFACTURER = "Seeed"
)

var (
	usb_VID uint16 = 0x2886
	usb_PID uint16 = 0x802F
)

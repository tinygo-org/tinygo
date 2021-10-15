// +build metro_m4_airlift

package machine

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xf01669ef

// GPIO Pins
const (
	D0 = PA23 // UART0 RX/PWM available
	D1 = PA22 // UART0 TX/PWM available
	D2 = PB17 // PWM available
	D3 = PB16 // PWM available
	D4 = PB13 // PWM available
	D5 = PB14 // PWM available
	D6 = PB15 // PWM available
	D7 = PB12 // PWM available

	D8  = PA21 // PWM available
	D9  = PA20 // PWM available
	D10 = PA18 // can be used for PWM or UART1 TX
	D11 = PA19 // can be used for PWM or UART1 RX
	D12 = PA17 // PWM available
	D13 = PA16 // PWM available

	D40 = PB22 // built-in neopixel
)

// Analog pins
const (
	A0 = PA02 // ADC/AIN[0]
	A1 = PA05 // ADC/AIN[2]
	A2 = PB06 // ADC/AIN[3]
	A3 = PB00 // ADC/AIN[4] // NOTE: different between "airlift" and non-airlift versions
	A4 = PB08 // ADC/AIN[5]
	A5 = PB09 // ADC/AIN[10]
)

const (
	LED    = D13
	WS2812 = D40
)

// USBCDC pins
const (
	USBCDC_DM_PIN = PA24
	USBCDC_DP_PIN = PA25
)

const (
	UART_TX_PIN = D1
	UART_RX_PIN = D0
)

const (
	UART2_TX_PIN = PA04
	UART2_RX_PIN = PA07
)

var (
	UART1 = &sercomUSART3
	UART2 = &sercomUSART0

	DefaultUART = UART1
)

const (
	NINA_CS     = PA15
	NINA_ACK    = PB04
	NINA_GPIO0  = PB01
	NINA_RESETN = PB05

	NINA_TX  = PA04
	NINA_RX  = PA07
	NINA_RTS = PB23
)

// I2C pins
const (
	SDA_PIN = PB02 // SDA: SERCOM5/PAD[0]
	SCL_PIN = PB03 // SCL: SERCOM5/PAD[1]
)

// SPI pins
const (
	SPI0_SCK_PIN = PA13 // SCK:  SERCOM2/PAD[1]
	SPI0_SDO_PIN = PA12 // SDO: SERCOM2/PAD[0]
	SPI0_SDI_PIN = PA14 // SDI: SERCOM2/PAD[2]

	NINA_SDO = SPI0_SDO_PIN
	NINA_SDI = SPI0_SDI_PIN
	NINA_SCK = SPI0_SCK_PIN
)

const (
	SPI1_SCK_PIN = D12 // SDI: SERCOM1/PAD[1]
	SPI1_SDO_PIN = D11 // SDO: SERCOM1/PAD[3]
	SPI1_SDI_PIN = D13 // SCK:  SERCOM1/PAD[0]
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit Metro M4 Airlift Lite"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x8037
)

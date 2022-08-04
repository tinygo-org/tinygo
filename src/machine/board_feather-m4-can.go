//go:build feather_m4_can
// +build feather_m4_can

package machine

import (
	"tinygo.org/x/device/sam"
)

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xf01669ef

// GPIO Pins
const (
	D0  = PB17 // UART0 RX/PWM available
	D1  = PB16 // UART0 TX/PWM available
	D4  = PA14 // PWM available
	D5  = PA16 // PWM available
	D6  = PA18 // PWM available
	D7  = PB03 // neopixel power
	D8  = PB02 // built-in neopixel
	D9  = PA19 // PWM available
	D10 = PA20 // can be used for PWM or UART1 TX
	D11 = PA21 // can be used for PWM or UART1 RX
	D12 = PA22 // PWM available
	D13 = PA23 // PWM available
	D21 = PA13 // PWM available
	D22 = PA12 // PWM available
	D23 = PB22 // PWM available
	D24 = PB23 // PWM available
	D25 = PA17 // PWM available
)

// Analog pins
const (
	A0 = PA02 // ADC/AIN[0]
	A1 = PA05 // ADC/AIN[2]
	A2 = PB08 // ADC/AIN[3]
	A3 = PB09 // ADC/AIN[4]
	A4 = PA04 // ADC/AIN[5]
	A5 = PA06 // ADC/AIN[10]
)

const (
	LED       = D13
	NEOPIXELS = D8
	WS2812    = D8
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
	UART2_TX_PIN = A4
	UART2_RX_PIN = A5
)

// I2C pins
const (
	SDA_PIN = D22 // SDA: SERCOM2/PAD[0]
	SCL_PIN = D21 // SCL: SERCOM2/PAD[1]
)

// SPI pins
const (
	SPI0_SCK_PIN = D25 // SCK: SERCOM1/PAD[1]
	SPI0_SDO_PIN = D24 // SDO: SERCOM1/PAD[3]
	SPI0_SDI_PIN = D23 // SDI: SERCOM1/PAD[2]
)

// CAN pins
const (
	CAN0_TX = PA22
	CAN0_RX = PA23

	CAN1_STANDBY = PB12
	CAN1_TX      = PB14
	CAN1_RX      = PB15
	BOOST_EN     = PB13 // power control of CAN1's TCAN1051HGV (H: enable)

	CAN_STANDBY = CAN1_STANDBY
	CAN_S       = CAN1_STANDBY
	CAN_TX      = CAN1_TX
	CAN_RX      = CAN1_RX
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit Feather M4 CAN"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x80CD
)

var (
	UART1 = &sercomUSART5

	UART2 = &sercomUSART0
)

func init() {
	// turn on neopixel
	D7.Configure(PinConfig{Mode: PinOutput})
	D7.High()
}

// I2C on the Feather M4 CAN.
var (
	I2C0 = sercomI2CM2
)

// SPI on the Feather M4 CAN.
var SPI0 = sercomSPIM1

// CAN on the Feather M4 CAN.
var (
	CAN0 = CAN{
		Bus: sam.CAN0,
	}

	CAN1 = CAN{
		Bus: sam.CAN1,
	}
)

var (
	DefaultUART = UART1
)

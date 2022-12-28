//go:build circuitplay_bluefruit
// +build circuitplay_bluefruit

package machine

const HasLowFrequencyCrystal = false

// GPIO Pins
const (
	D0  = P0_30
	D1  = P0_14
	D2  = P0_05
	D3  = P0_04
	D4  = P1_02
	D5  = P1_15
	D6  = P0_02
	D7  = P1_06
	D8  = P0_13
	D9  = P0_29
	D10 = P0_03
	D11 = P1_04
	D12 = P0_26
	D13 = P1_14
)

// Analog Pins
const (
	A1 = P0_02
	A2 = P0_29
	A3 = P0_03
	A4 = P0_04
	A5 = P0_05
	A6 = P0_30
	A7 = P0_14
	A8 = P0_28
	A9 = P0_31
)

const (
	LED       = D13
	NEOPIXELS = D8
	WS2812    = D8

	BUTTONA = D4
	BUTTONB = D5
	SLIDER  = D7 // built-in slide switch

	BUTTON  = BUTTONA
	BUTTON1 = BUTTONB

	LIGHTSENSOR = A8
	TEMPSENSOR  = A9
)

// UART0 pins (logical UART1)
const (
	UART_TX_PIN = P0_14 // PORTB
	UART_RX_PIN = P0_30 // PORTB
)

// I2C pins
const (
	SDA_PIN = P0_05 // I2C0 external
	SCL_PIN = P0_04 // I2C0 external

	SDA1_PIN = P1_10 // I2C1 internal
	SCL1_PIN = P1_12 // I2C1 internal
)

// SPI pins (internal flash)
const (
	SPI0_SCK_PIN = P0_19 // SCK
	SPI0_SDO_PIN = P0_21 // SDO
	SPI0_SDI_PIN = P0_23 // SDI
)

// PDM pins
const (
	PDM_CLK_PIN = P0_17 // CLK
	PDM_DIN_PIN = P0_16 // DIN
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit Circuit Playground Bluefruit"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x8045
)

var (
	DefaultUART = UART0
)

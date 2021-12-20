// +build sam,atsamd51,pygamer

package machine

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xf01669ef

// GPIO Pins
const (
	D0  = PB17 // UART0 RX/PWM available
	D1  = PB16 // UART0 TX/PWM available
	D2  = PB03
	D3  = PB02
	D4  = PA14 // PWM available
	D5  = PA16 // PWM available
	D6  = PA18 // PWM available
	D7  = PB14 // CS for microSD card slot
	D8  = PA15 // built-in neopixel
	D9  = PA19 // PWM available
	D10 = PA20 // can be used for PWM or UART1 TX
	D11 = PA21 // can be used for PWM or UART1 RX
	D12 = PA22 // PWM available
	D13 = PA23 // PWM available
)

// Analog pins
const (
	A0 = PA02 // ADC/AIN[0]
	A1 = PA05 // ADC/AIN[2]
	A2 = PB08 // ADC/AIN[3]
	A3 = PB09 // ADC/AIN[4]
	A4 = PA04 // ADC/AIN[5]
	A5 = PA06 // ADC/AIN[6]
	A6 = PB01 // ADC/AIN[12]/VMEAS
	A7 = PB04 // ADC/AIN[6]/LIGHT
	A8 = D2   // ADC/AIN[14]
	A9 = D3   // ADC/AIN[15]
)

const (
	LED       = D13
	NEOPIXELS = D8
	WS2812    = D8

	SD_CS = D7

	LIGHTSENSOR = A7

	BUTTON_LATCH = PB00
	BUTTON_OUT   = PB30
	BUTTON_CLK   = PB31

	JOYY = PB06
	JOYX = PB07

	TFT_DC   = PB05
	TFT_CS   = PB12
	TFT_RST  = PA00
	TFT_LITE = PA01

	SPEAKER_ENABLE = PA27
)

const (
	BUTTON_SELECT_MASK = 16
	BUTTON_START_MASK  = 32
	BUTTON_A_MASK      = 64
	BUTTON_B_MASK      = 128
)

// USBCDC pins
const (
	USBCDC_DM_PIN = PA24
	USBCDC_DP_PIN = PA25
)

// UART1 pins
const (
	UART_TX_PIN = D1
	UART_RX_PIN = D0
)

// UART1 var is on SERCOM3, defined in atsamd51.go

// UART2 pins
const (
	UART2_TX_PIN = A4
	UART2_RX_PIN = A5
)

// UART2 var is on SERCOM0, defined in atsamd51.go

// I2C pins
const (
	SDA_PIN = PA12 // SDA: SERCOM2/PAD[0]
	SCL_PIN = PA13 // SCL: SERCOM2/PAD[1]
)

// I2C on the PyGamer.
var (
	I2C0 = sercomI2CM2
)

// SPI pins
const (
	SPI0_SCK_PIN = PA17 // SCK: SERCOM1/PAD[1]
	SPI0_SDO_PIN = PB23 // SDO: SERCOM1/PAD[3]
	SPI0_SDI_PIN = PB22 // SDI: SERCOM1/PAD[2]
)

// SPI on the PyGamer.
var SPI0 = sercomSPIM1

// TFT SPI pins
const (
	SPI1_SCK_PIN = PB13 // SCK: SERCOM4/PAD[1]
	SPI1_SDO_PIN = PB15 // SDO: SERCOM4/PAD[3]
	SPI1_SDI_PIN = NoPin
)

// TFT SPI on the PyGamer.
var SPI1 = sercomSPIM4

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit pyGamer M4"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x8033
)

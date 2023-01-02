//go:build sam && atsamd21 && feather_m0_express

package machine

// used to reset into bootloader
const resetMagicValue = 0xf01669ef

// GPIO Pins
const (
	D0  = PA11  // UART0 RX
	D1  = PA10  // UART0 TX
	D2  = NoPin // does not seem to exist
	D3  = NoPin // does not seem to exist
	D4  = NoPin // does not seem to exist
	D5  = PA15
	D6  = PA20
	D7  = NoPin // does not seem to exist
	D8  = PA06  // NEOPIXEL
	D9  = PA07
	D10 = PA18
	D11 = PA16
	D12 = PA19
	D13 = PA17
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
	LED      = D13
	NEOPIXEL = D8
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

	UART1_TX_PIN = D10
	UART1_RX_PIN = D12
)

// UART0 on the Feather M0 Express.
var UART0 = &sercomUSART0
var UART1 = &sercomUSART1

// I2C pins
const (
	SDA_PIN = PA22 // SDA: SERCOM3/PAD[0]
	SCL_PIN = PA23 // SCL: SERCOM3/PAD[1]
)

// I2C on the Feather M0 Express.
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
	I2S_SD_PIN  = PA07
	I2S_WS_PIN  = NoPin // TODO: figure out what this is on Feather M0 Express.
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
	DefaultUART = UART0
)

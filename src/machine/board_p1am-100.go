// +build p1am_100

// This contains the pin mappings for the ProductivityOpen P1AM-100 board.
//
// For more information, see: https://facts-engineering.github.io/
//
package machine

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0x07738135

// Note: On the P1AM-100, pins D8, D9, D10, A3, and A4 are used for
// communication with the base controller.

// GPIO Pins
const (
	D0 Pin = PA22 // PWM available
	D1 Pin = PA23 // PWM available
	D2 Pin = PA10 // PWM available
	D3 Pin = PA11 // PWM available
	D4 Pin = PB10 // PWM available
	D5 Pin = PB11 // PWM available
	D6 Pin = PA20 // PWM available
	D7 Pin = PA21 // PWM available

	D8  Pin = PA16 // PWM available
	D9  Pin = PA17
	D10 Pin = PA19 // PWM available
	D11 Pin = PA08
	D12 Pin = PA09
	D13 Pin = PB23
	D14 Pin = PB22

	// Remaining pins are shared with analog pins
	D15 Pin = PA02

	D16 Pin = PB02
	D17 Pin = PB03
	D18 Pin = PA04 // PWM available
	D19 Pin = PA05 // PWM available
	D20 Pin = PA06
	D21 Pin = PA07
)

// Analog pins
const (
	A0 Pin = PA02 // ADC/AIN[0]
	A1 Pin = PB02 // ADC/AIN[10]
	A2 Pin = PB03 // ADC/AIN[11]
	A3 Pin = PA04 // ADC/AIN[4]
	A4 Pin = PA05 // ADC/AIN[5]
	A5 Pin = PA06 // ADC/AIN[6]
	A6 Pin = PA07 // ADC/AIN[7]
)

const (
	SWITCH      Pin = PA28
	LED         Pin = PB08
	ADC_BATTERY Pin = PB09 // ADC/AIN[3]
)

// P1AM Base Controller
const (
	BASE_SLAVE_SELECT_PIN Pin = A3
	BASE_SLAVE_ACK_PIN    Pin = A4
	BASE_ENABLE_PIN       Pin = PB09
)

// USBCDC pins
const (
	USBCDC_DM_PIN          Pin = PA24
	USBCDC_DP_PIN          Pin = PA25
	USBCDC_HOST_ENABLE_PIN Pin = PA18
)

// UART1 pins
const (
	UART_RX_PIN Pin = PB23 // RX: SERCOM5/PAD[3]
	UART_TX_PIN Pin = PB22 // TX: SERCOM5/PAD[2]
)

// UART1 on the P1AM-100 connects to the normal TX/RX pins.
var UART1 = &sercomUSART5

// I2C pins
const (
	SDA_PIN Pin = PA08 // SDA:  SERCOM0/PAD[0]
	SCL_PIN Pin = PA09 // SCL:  SERCOM0/PAD[1]
)

// I2C on the P1AM-100.
var (
	I2C0 = sercomI2CM0
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = D9  // SCK: SERCOM1/PAD[1]
	SPI0_SDO_PIN Pin = D8  // SDO: SERCOM1/PAD[0]
	SPI0_SDI_PIN Pin = D10 // SDI: SERCOM1/PAD[3]
)

// SD card pins
const (
	SDCARD_SDI_PIN Pin = PA15 // SDI: SERCOM2/PAD[3]
	SDCARD_SDO_PIN Pin = PA12 // SDO: SERCOM2/PAD[0]
	SDCARD_SCK_PIN Pin = PA13 // SCK: SERCOM2/PAD[1]
	SDCARD_SS_PIN  Pin = PA14 // SS: as GPIO
	SDCARD_CD_PIN  Pin = PA27
)

// I2S pins
const (
	I2S_SCK_PIN Pin = D2
	I2S_SD_PIN  Pin = A6
	I2S_WS_PIN      = D3
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "P1AM-100"
	usb_STRING_MANUFACTURER = "Facts Engineering"
)

var (
	usb_VID uint16 = 0x1354
	usb_PID uint16 = 0x4000
)

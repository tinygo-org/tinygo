// +build pyportal

package machine

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0xf01669ef

// GPIO Pins
const (
	D0  = PB13 // NINA_RX
	D1  = PB12 // NINA_TX
	D2  = PB22 // built-in neopixel
	D3  = PA04 // PWM available
	D4  = PA05 // PWM available
	D5  = PB16 // NINA_ACK
	D6  = PB15 // NINA_GPIO0
	D7  = PB17 // NINA_RESETN
	D8  = PB14 // NINA_CS
	D9  = PB04 // TFT_RD
	D10 = PB05 // TFT_DC
	D11 = PB06 // TFT_CS
	D12 = PB07 // TFT_TE
	D13 = PB23 // built-in LED
	D24 = PA00 // TFT_RESET
	D25 = PB31 // TFT_BACKLIGHT
	D26 = PB09 // TFT_WR
	D27 = PB02 // SDA
	D28 = PB03 // SCL
	D29 = PA12 // SDO
	D30 = PA13 // SCK
	D31 = PA14 // SDI
	D32 = PB30 // SD_CS
	D33 = PA01 // SD_CARD_DETECT
	D34 = PA16 // LCD_DATA0
	D35 = PA17 // LCD_DATA1
	D36 = PA18 // LCD_DATA2
	D37 = PA19 // LCD_DATA3
	D38 = PA20 // LCD_DATA4
	D39 = PA21 // LCD_DATA5
	D40 = PA22 // LCD_DATA6
	D41 = PA23 // LCD_DATA7
	D42 = PB10 // QSPI
	D43 = PB11 // QSPI
	D44 = PA08 // QSPI
	D45 = PA09 // QSPI
	D46 = PA10 // QSPI
	D47 = PA11 // QSPI
	D50 = PA02 // speaker amplifier shutdown
	D51 = PA15 // NINA_RTS

	NINA_CS     = D8
	NINA_ACK    = D5
	NINA_GPIO0  = D6
	NINA_RESETN = D7

	NINA_TX  = D1
	NINA_RX  = D0
	NINA_RTS = D51

	LCD_DATA0 = D34

	TFT_RD        = D9
	TFT_DC        = D10
	TFT_CS        = D11
	TFT_TE        = D12
	TFT_RESET     = D24
	TFT_BACKLIGHT = D25
	TFT_WR        = D26

	NEOPIXEL = D2
	WS2812   = D2
	SPK_SD   = D50
)

// Analog pins
const (
	A0 = PA02 // ADC0/AIN[0]
	A1 = D3   // ADC0/AIN[4]
	A2 = PA07 // ADC0/AIN[7]
	A3 = D4   // ADC0/AIN[5]
	A4 = PB00 // ADC0/AIN[12]
	A5 = PB01 // ADC0/AIN[13]
	A6 = PA06 // ADC0/AIN[6]
	A7 = PB08 // ADC1/AIN[0]

	AUDIO_OUT = A0
	LIGHT     = A2
	TOUCH_YD  = A4
	TOUCH_XL  = A5
	TOUCH_YU  = A6
	TOUCH_XR  = A7
)

const (
	LED = D13
)

// USBCDC pins
const (
	USBCDC_DM_PIN = PA24
	USBCDC_DP_PIN = PA25
)

// UART1 aka NINA_TX/NINA_RX
const (
	UART_TX_PIN = D1
	UART_RX_PIN = D0
)

var (
	UART1 = &sercomUSART4

	DefaultUART = UART1
)

// I2C pins
const (
	SDA_PIN = PB02 // SDA: SERCOM2/PAD[0]
	SCL_PIN = PB03 // SCL: SERCOM2/PAD[1]
)

// I2C on the PyPortal.
var (
	I2C0 = sercomI2CM5
)

// SPI pins
const (
	SPI0_SCK_PIN = PA13 // SCK: SERCOM1/PAD[1]
	SPI0_SDO_PIN = PA12 // SDO: SERCOM1/PAD[3]
	SPI0_SDI_PIN = PA14 // SDI: SERCOM1/PAD[2]

	NINA_SDO = SPI0_SDO_PIN
	NINA_SDI = SPI0_SDI_PIN
	NINA_SCK = SPI0_SCK_PIN
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit PyPortal M4"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x8035
)

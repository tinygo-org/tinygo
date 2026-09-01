//go:build sam && atsamd21 && arduino_m0_pro

package machine

// used to reset into bootloader
const resetMagicValue = 0x07738135

// GPIO Pins - Digital Low
const (
	D0 = PA11 // RX
	D1 = PA10 // TX
	D2 = PA08
	D3 = PA09 // PWM available
	D4 = PA14 // PWM available
	D5 = PA15 // PWM available
	D6 = PA20 // PWM available
	D7 = PA21
)

// GPIO Pins - Digital High
const (
	D8  = PA06 // PWM available
	D9  = PA07 // PWM available
	D10 = PA18 // PWM available
	D11 = PA16 // PWM available
	D12 = PA19 // PWM available
	D13 = PA17 // PWM available
)

// ADC pins
const (
	AREF Pin = PA03
	ADC0 Pin = PA02
	ADC1 Pin = PB08
	ADC2 Pin = PB09
	ADC3 Pin = PA04
	ADC4 Pin = PA05
	ADC5 Pin = PB02
)

// LEDs on the Arduino M0 Pro
const (
	LED      = LED1
	LED1 Pin = D13
	LED2 Pin = PA27 // TX LED
	LED3 Pin = PB03 // RX LED
)

// SPI pins
const (
	SPI0_SDO_PIN Pin = PB10 // MOSI
	SPI0_SDI_PIN Pin = PB12 // MISO
	SPI0_SCK_PIN Pin = PB11 // SCK
)

// I2C pins
const (
	SDA_PIN Pin = PA22 // SDA: Pin 20
	SCL_PIN Pin = PA23 // SCL: Pin 21
)

// I2S pins - might not be exposed
const (
	I2S_SCK_PIN Pin = PA10
	I2S_SDO_PIN Pin = PA07
	I2S_SDI_PIN     = NoPin
	I2S_WS_PIN  Pin = PA11
)

// UART pins
const (
	UART_RX_PIN Pin = D0
	UART_TX_PIN Pin = D1
)

// 'native' USB port pins
const (
	USBCDC_DM_PIN Pin = PA24
	USBCDC_DP_PIN Pin = PA25
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Arduino M0 Pro"
	usb_STRING_MANUFACTURER = "Arduino LLC"

	usb_VID uint16 = 0x2A03
	usb_PID uint16 = 0x804F
)

// 32.768 KHz Crystal
const (
	XIN32  Pin = PA00
	XOUT32 Pin = PA01
)

// +build sam,atsamd21,arduino_zero

package machine

// used to reset into bootloader
const RESET_MAGIC_VALUE = 0x07738135

// GPIO Pins
const (
	D0  = PA11 // RX
	D1  = PA10 // TX
	D2  = PA14
	D3  = PA09 // PWM available
	D4  = PA08 // PWM available
	D5  = PA15 // PWM available
	D6  = PA20 // PWM available
	D7  = PA21
	D8  = PA06 // PWM available
	D9  = PA07 // PWM available
	D10 = PA18 // PWM available
	D11 = PA16 // PWM available
	D12 = PA19 // PWM available
	D13 = PA17 // PWM available
)

// LEDs on the Arduino Zero
const (
	LED  Pin = D13
	LED1 Pin = PA27
	LED2 Pin = PB03
)

// ADC on the Arduino
const (
	ADC0 Pin = PA02
	ADC1 Pin = PB08
	ADC2 Pin = PB09
	ADC3 Pin = PA04
	ADC4 Pin = PA05
	ADC5 Pin = PB02
)

// SPI pins
const (
	SPI0_SDO_PIN = PA16 // MOSI: SERCOM1/PAD[0]
	SPI0_SCK_PIN = PA17 // SCK:  SERCOM1/PAD[1]
	SPI0_SS_Pin  = PA18 // SS:   SERCOM1/PAD[2]
	SPI0_SDI_PIN = PA19 // MISO: SERCOM1/PAD[3]
)

// I2C pins
const (
	SDA_PIN Pin = PA22 // SDA: SERCOM3/PAD[0]
	SCL_PIN Pin = PA23 // SCL: SERCOM3/PAD[1]
)

// I2S pins
const (
	I2S_SCK_PIN Pin = PA10
	I2S_SD_PIN  Pin = PA07
	I2S_WS_PIN  Pin = PA11
)

// UART0 pins - 'promgramming' USB port
const (
	UART_TX_PIN Pin = D1
	UART_RX_PIN Pin = D0
)

// 'native' USB port pins
const (
	USBCDC_DM_PIN Pin = PA25
	USBCDC_DP_PIN Pin = PA24
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Arduino Zero"
	usb_STRING_MANUFACTURER = "Arduino LLC"
)

var (
	usb_VID uint16 = 0x03eb
	usb_PID uint16 = 0x2157
)

//go:build bluemicro840

package machine

const HasLowFrequencyCrystal = true

// GPIO Pins
const (
	D006 = P0_06
	D008 = P0_08
	D015 = P0_15
	D017 = P0_17
	D020 = P0_20
	D013 = P0_13
	D024 = P0_24
	D009 = P0_09
	D010 = P0_10
	D106 = P1_06

	D031 = P0_31 // AIN7; P0.31 (AIN7) is used to read the voltage of the battery via ADC. It can’t be used for any other function.
	D012 = P0_12 // VCC 3.3V; P0.12 on VCC shuts off the power to VCC when you set it to high; This saves on battery immensely for LEDs of all kinds that eat power even when off

	D030 = P0_30
	D026 = P0_26
	D029 = P0_29
	D002 = P0_02
	D113 = P1_13
	D003 = P0_03
	D028 = P0_28
	D111 = P1_11
)

// Analog Pins
const (
	AIN0 = P0_02
	AIN1 = P0_03
	AIN2 = P0_04 // Not Connected
	AIN3 = P0_05 // Not Connected
	AIN4 = P0_28
	AIN5 = P0_29
	AIN6 = P0_30
	AIN7 = P0_31 // Battery
)

const (
	LED1 Pin = P1_04 // Red LED
	LED2 Pin = P1_10 // Blue LED
	LED  Pin = LED1
)

// UART0 pins (logical UART1) - Maps to same location as Pro Micro
const (
	UART_RX_PIN = P0_08
	UART_TX_PIN = P0_06
)

// I2C pins
const (
	SDA_PIN = P0_15 // I2C0 external
	SCL_PIN = P0_17 // I2C0 external
)

// SPI pins
const (
	SPI0_SCK_PIN = P1_13 // SCK
	SPI0_SDI_PIN = P0_03 // SDI
	SPI0_SDO_PIN = P0_28 // SDO

)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "bluemicro840"
	usb_STRING_MANUFACTURER = "BlueMicro"
)

var (
	usb_VID uint16 = 0x1d50
	usb_PID uint16 = 0x6161
)

//go:build nicenano

package machine

const HasLowFrequencyCrystal = true

// GPIO Pins
const (
	D006 = P0_06
	D008 = P0_08
	D017 = P0_17
	D020 = P0_20
	D022 = P0_22
	D024 = P0_24
	D100 = P1_00
	D011 = P0_11
	D104 = P1_04
	D106 = P1_06

	D004 = P0_04 // AIN2; P0.04 (AIN2) is used to read the voltage of the battery via ADC. It can’t be used for any other function.
	D013 = P0_13 // VCC 3.3V; P0.13 on VCC shuts off the power to VCC when you set it to high; This saves on battery immensely for LEDs of all kinds that eat power even when off
	D115 = P1_15
	D113 = P1_13
	D031 = P0_31 // AIN7
	D029 = P0_29 // AIN5
	D002 = P0_02 // AIN0

	D111 = P1_11
	D010 = P0_10 // NFC2
	D009 = P0_09 // NFC1

	D026 = P0_26
	D012 = P0_12
	D101 = P1_01
	D102 = P1_02
	D107 = P1_07
)

// Analog Pins
const (
	AIN2 = P0_04 // Battery
	AIN7 = P0_31
	AIN5 = P0_29
	AIN0 = P0_02
)

const (
	LED = P0_15
)

// UART0 pins (logical UART1)
const (
	UART_RX_PIN = P0_06
	UART_TX_PIN = P0_08
)

// I2C pins
const (
	SDA_PIN = P0_17 // I2C0 external
	SCL_PIN = P0_20 // I2C0 external
)

// SPI pins
const (
	SPI0_SCK_PIN = P0_22 // SCK
	SPI0_SDO_PIN = P0_24 // SDO
	SPI0_SDI_PIN = P1_00 // SDI
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "nice!nano"
	usb_STRING_MANUFACTURER = "Nice Keyboards"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x0029
)

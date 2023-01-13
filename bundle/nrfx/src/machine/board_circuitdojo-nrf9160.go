//go:build feather_nrf9160
// +build feather_nrf9160

package machine

const HasLowFrequencyCrystal = true

// GPIO Pins
const (
	D0  = P0_26 // SDA
	D1  = P0_27 // SCL
	D2  = P0_29
	D3  = P0_30
	D4  = P0_00
	D5  = P0_01
	D6  = P0_02
	D7  = P0_03 // BlueLED
	D8  = P0_04
	D9  = P0_13 // A0
	D10 = P0_14 // A1
	D11 = P0_15 // A2
	D12 = P0_16 // A3
	D13 = P0_17 // A4
	D14 = P0_18 // A5
	D15 = P0_19 // SCK
	D16 = P0_20 // VBAT
	D17 = P0_21 // COPI
	D18 = P0_22 // CIPO
	D19 = P0_23 // RX
	D20 = P0_24 // TX
	D21 = P0_25 // VBAT_EN
	D22 = P0_12 // MD
	D23 = P0_05 // UART RX
	D24 = P0_06 // UART TX
	D25 = P0_07 // FLASH_CS
	D26 = P0_08 // FLASH_WP
	D27 = P0_09 // FLASH_SI
	D28 = P0_10 // FLASH_HOLD
	D29 = P0_11 // FLASH_SCK
	D30 = P0_28 // FLASH_SO
	D31 = P0_31
)

// Analog Pins
const (
	A0 = D9
	A1 = D10
	A2 = D11
	A3 = D12
	A4 = D13
	A5 = D14
	A6 = D16 // Battery
)

const (
	LED      = D7
	LED1     = LED
	BUTTON   = D22
)

// UART0 pins
const (
	UART_RX_PIN = D23
	UART_TX_PIN = D24
)

// I2C pins
const (
	SDA_PIN = D0 // I2C0 external
	SCL_PIN = D1 // I2C0 external
)

// SPI pins
const (
	SPI0_SCK_PIN = D29 // SCK
	SPI0_SDO_PIN = D30 // SDO
	SPI0_SDI_PIN = D27 // SDI
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Feather nRF9160"
	usb_STRING_MANUFACTURER = "Cirtuitdojo"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x802A
)

var (
	DefaultUART = UART0
)

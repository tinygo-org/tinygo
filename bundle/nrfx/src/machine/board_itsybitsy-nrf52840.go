//go:build itsybitsy_nrf52840

package machine

const HasLowFrequencyCrystal = true

// GPIO Pins
const (
	D0  = P0_25 // UART TX
	D1  = P0_24 // UART RX
	D2  = P1_02
	D3  = P0_06 // LED1
	D4  = P0_29 // Button
	D5  = P0_27
	D6  = P1_09 // DotStar Clock
	D7  = P1_08
	D8  = P0_08 // DotStar Data
	D9  = P0_07
	D10 = P0_05
	D11 = P0_26
	D12 = P0_11
	D13 = P0_12
	D14 = P0_04 // A0
	D15 = P0_30 // A1
	D16 = P0_28 // A2
	D17 = P0_31 // A3
	D18 = P0_02 // A4
	D19 = P0_03 // A5
	D20 = P0_05 // A6
	D21 = P0_16 // I2C SDA
	D22 = P0_14 // I2C SCL
	D23 = P0_20 // SPI SDI
	D24 = P0_15 // SPI SDO
	D25 = P0_13 // SPI SCK
	D26 = P0_19 // QSPI SCK
	D27 = P0_23 // QSPI CS
	D28 = P0_21 // QSPI Data 0
	D29 = P0_22 // QSPI Data 1
	D30 = P1_00 // QSPI Data 2
	D31 = P0_17 // QSPI Data 3
)

// Analog Pins
const (
	A0 = D14
	A1 = D15
	A2 = D16
	A3 = D17
	A4 = D18
	A5 = D19
	A6 = D20
)

const (
	LED    = D3
	LED1   = LED
	BUTTON = D4

	QSPI_SCK   = D26
	QSPI_CS    = D27
	QSPI_DATA0 = D28
	QSPI_DATA1 = D29
	QSPI_DATA2 = D30
	QSPI_DATA3 = D31
)

// UART0 pins (logical UART1)
const (
	UART_RX_PIN = D0
	UART_TX_PIN = D1
)

// I2C pins
const (
	SDA_PIN = D21 // I2C0 external
	SCL_PIN = D22 // I2C0 external
)

// SPI pins
const (
	SPI0_SCK_PIN = D25
	SPI0_SDO_PIN = D24
	SPI0_SDI_PIN = D23
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit ItsyBitsy nRF52840 Express"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x8051
)

var (
	DefaultUART = UART0
)

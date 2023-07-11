//go:build feather_nrf52840

package machine

const HasLowFrequencyCrystal = true

// GPIO Pins
const (
	D0  = P0_25 // UART TX
	D1  = P0_24 // UART RX
	D2  = P0_10 // NFC2
	D3  = P1_15 // LED1
	D4  = P1_10 // LED2
	D5  = P1_08
	D6  = P0_07
	D7  = P1_02 // Button
	D8  = P0_16 // NeoPixel
	D9  = P0_26
	D10 = P0_27
	D11 = P0_06
	D12 = P0_08
	D13 = P1_09
	D14 = P0_04 // A0
	D15 = P0_05 // A1
	D16 = P0_30 // A2
	D17 = P0_28 // A3
	D18 = P0_02 // A4
	D19 = P0_03 // A5
	D20 = P0_29 // Battery
	D21 = P0_31 // AREF
	D22 = P0_12 // I2C SDA
	D23 = P0_11 // I2C SCL
	D24 = P0_15 // SPI MISO
	D25 = P0_13 // SPI MOSI
	D26 = P0_14 // SPI SCK
	D27 = P0_19 // QSPI CLK
	D28 = P0_20 // QSPI CS
	D29 = P0_17 // QSPI Data 0
	D30 = P0_22 // QSPI Data 1
	D31 = P0_23 // QSPI Data 2
	D32 = P0_21 // QSPI Data 3
	D33 = P0_09 // NFC1 (test point on bottom of board)
)

// Analog Pins
const (
	A0 = D14
	A1 = D15
	A2 = D16
	A3 = D17
	A4 = D18
	A5 = D19
	A6 = D20 // Battery
	A7 = D21 // ARef
)

const (
	LED      = D3
	LED1     = LED
	LED2     = D4
	NEOPIXEL = D8
	WS2812   = D8
	BUTTON   = D7

	QSPI_SCK   = D27
	QSPI_CS    = D28
	QSPI_DATA0 = D29
	QSPI_DATA1 = D30
	QSPI_DATA2 = D31
	QSPI_DATA3 = D32
)

// UART0 pins (logical UART1)
const (
	UART_RX_PIN = D1
	UART_TX_PIN = D0
)

// I2C pins
const (
	SDA_PIN = D22 // I2C0 external
	SCL_PIN = D23 // I2C0 external
)

// SPI pins
const (
	SPI0_SCK_PIN = D26 // SCK
	SPI0_SDO_PIN = D25 // SDO
	SPI0_SDI_PIN = D24 // SDI
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Feather nRF52840 Express"
	usb_STRING_MANUFACTURER = "Adafruit Industries LLC"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x802A
)

var (
	DefaultUART = UART0
)

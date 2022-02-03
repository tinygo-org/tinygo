//go:build clue_alpha
// +build clue_alpha

package machine

const HasLowFrequencyCrystal = false

// GPIO Pins
const (
	D0  = P0_04
	D1  = P0_05
	D2  = P0_03
	D3  = P0_28
	D4  = P0_02
	D5  = P1_02
	D6  = P1_09
	D7  = P0_07
	D8  = P1_07
	D9  = P0_27
	D10 = P0_30
	D11 = P1_10
	D12 = P0_31
	D13 = P0_08
	D14 = P0_06
	D15 = P0_26
	D16 = P0_29
	D17 = P1_01
	D18 = P0_16
	D19 = P0_25
	D20 = P0_24
	D21 = A0
	D22 = A1
	D23 = A2
	D24 = A3
	D25 = A4
	D26 = A5
	D27 = A6
	D28 = A7
	D29 = P0_14
	D30 = P0_15
	D31 = P0_12
	D32 = P0_13
	D33 = P1_03
	D34 = P1_05
	D35 = P0_00
	D36 = P0_01
	D37 = P0_19
	D38 = P0_20
	D39 = P0_17
	D40 = P0_22
	D41 = P0_23
	D42 = P0_21
	D43 = P0_10
	D44 = P0_09
	D45 = P1_06
	D46 = P1_00
)

// Analog Pins
const (
	A0 = D12
	A1 = D16
	A2 = D0
	A3 = D1
	A4 = D2
	A5 = D3
	A6 = D4
	A7 = D10
)

const (
	LED      = D17
	LED1     = LED
	LED2     = D43
	NEOPIXEL = D18
	WS2812   = D18

	BUTTON_LEFT  = D5
	BUTTON_RIGHT = D11

	// 240x240 ST7789 display is connected to these pins (use RowOffset = 80)
	TFT_SCK   = D29
	TFT_SDO   = D30
	TFT_CS    = D31
	TFT_DC    = D32
	TFT_RESET = D33
	TFT_LITE  = D34

	PDM_DAT = D35
	PDM_CLK = D36

	QSPI_SCK   = D37
	QSPI_CS    = D38
	QSPI_DATA0 = D39
	QSPI_DATA1 = D40
	QSPI_DATA2 = D41
	QSPI_DATA3 = D42

	SPEAKER = D46
)

// UART0 pins (logical UART1)
const (
	UART_RX_PIN = D0
	UART_TX_PIN = D1
)

// I2C pins
const (
	SDA_PIN = D20 // I2C0 external
	SCL_PIN = D19 // I2C0 external
)

// SPI pins
const (
	SPI0_SCK_PIN = D13 // SCK
	SPI0_SDO_PIN = D15 // SDO
	SPI0_SDI_PIN = D14 // SDI
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "Adafruit CLUE"
	usb_STRING_MANUFACTURER = "Adafruit"
)

var (
	usb_VID uint16 = 0x239A
	usb_PID uint16 = 0x8072
)

var (
	DefaultUART = UART0
)

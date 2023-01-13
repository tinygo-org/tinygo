//go:build microbit_v2

package machine

// The micro:bit does not have a 32kHz crystal on board.
const HasLowFrequencyCrystal = false

// Buttons on the micro:bit v2 (A and B)
const (
	BUTTON  Pin = BUTTONA
	BUTTONA Pin = P5
	BUTTONB Pin = P11
)

var DefaultUART = UART0

// UART pins
const (
	UART_TX_PIN Pin = P34
	UART_RX_PIN Pin = P33
)

// ADC pins
const (
	ADC0 Pin = P0
	ADC1 Pin = P1
	ADC2 Pin = P2
)

// I2C0 (internal) pins
const (
	SDA_PIN  Pin = SDA0_PIN
	SCL_PIN  Pin = SCL0_PIN
	SDA0_PIN Pin = P30
	SCL0_PIN Pin = P31
)

// I2C1 (external) pins
const (
	SDA1_PIN Pin = P20
	SCL1_PIN Pin = P19
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = P13
	SPI0_SDO_PIN Pin = P15
	SPI0_SDI_PIN Pin = P14
)

// GPIO/Analog pins
const (
	P0  Pin = 2
	P1  Pin = 3
	P2  Pin = 4
	P3  Pin = 31
	P4  Pin = 28
	P5  Pin = 14
	P6  Pin = 37
	P7  Pin = 11
	P8  Pin = 10
	P9  Pin = 9
	P10 Pin = 30
	P11 Pin = 23
	P12 Pin = 12
	P13 Pin = 17
	P14 Pin = 1
	P15 Pin = 13
	P16 Pin = 34
	P19 Pin = 26
	P20 Pin = 32
	P21 Pin = 21
	P22 Pin = 22
	P23 Pin = 15
	P24 Pin = 24
	P25 Pin = 19
	P26 Pin = 36
	P27 Pin = 0
	P28 Pin = 20
	P29 Pin = 5
	P30 Pin = 16
	P31 Pin = 8
	P32 Pin = 25
	P33 Pin = 40
	P34 Pin = 6
)

// LED matrix pins
const (
	LED_COL_1 Pin = P0_28
	LED_COL_2 Pin = P0_11
	LED_COL_3 Pin = P0_31
	LED_COL_4 Pin = P1_05
	LED_COL_5 Pin = P0_30
	LED_ROW_1 Pin = P0_21
	LED_ROW_2 Pin = P0_22
	LED_ROW_3 Pin = P0_15
	LED_ROW_4 Pin = P0_24
	LED_ROW_5 Pin = P0_19
)

// Peripherals
const (
	BUZZER    = P27
	CAP_TOUCH = P26
	MIC       = P29
	MIC_LED   = P28
)

// USB CDC identifiers
const (
	usb_STRING_PRODUCT      = "BBC micro:bit V2"
	usb_STRING_MANUFACTURER = "BBC"
)

var (
	usb_VID uint16 = 0x0d28
	usb_PID uint16 = 0x0204
)

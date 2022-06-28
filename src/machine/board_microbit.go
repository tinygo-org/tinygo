//go:build microbit
// +build microbit

package machine

// The micro:bit does not have a 32kHz crystal on board.
const HasLowFrequencyCrystal = false

// Buttons on the micro:bit (A and B)
const (
	BUTTON  = BUTTONA
	BUTTONA = P0_17
	BUTTONB = P0_26
)

var DefaultUART = UART0

// UART pins
const (
	UART_TX_PIN = P0_24
	UART_RX_PIN = P0_25
)

// ADC pins
const (
	ADC0 = P0_03 // P0 on the board
	ADC1 = P0_02 // P1 on the board
	ADC2 = P0_01 // P2 on the board
)

// I2C pins
const (
	SDA_PIN = P0_30 // P20 on the board
	SCL_PIN = P0_00 // P19 on the board
)

// SPI pins
const (
	SPI0_SCK_PIN = P0_23 // P13 on the board
	SPI0_SDO_PIN = P0_21 // P15 on the board
	SPI0_SDI_PIN = P0_22 // P14 on the board
)

// GPIO/Analog pins
const (
	P0  = P0_03
	P1  = P0_02
	P2  = P0_01
	P3  = P0_04
	P4  = P0_05
	P5  = P0_17
	P6  = P0_12
	P7  = P0_11
	P8  = P0_18
	P9  = P0_10
	P10 = P0_06
	P11 = P0_26
	P12 = P0_20
	P13 = P0_23
	P14 = P0_22
	P15 = P0_21
	P16 = P0_16
)

// LED matrix pins
const (
	LED_COL_1 = P0_04
	LED_COL_2 = P0_05
	LED_COL_3 = P0_06
	LED_COL_4 = P0_07
	LED_COL_5 = P0_08
	LED_COL_6 = P0_09
	LED_COL_7 = P0_10
	LED_COL_8 = P0_11
	LED_COL_9 = P0_12
	LED_ROW_1 = P0_13
	LED_ROW_2 = P0_14
	LED_ROW_3 = P0_15
)

//go:build microbit
// +build microbit

package machine

// The micro:bit does not have a 32kHz crystal on board.
const HasLowFrequencyCrystal = false

// Buttons on the micro:bit (A and B)
const (
	BUTTON  Pin = BUTTONA
	BUTTONA Pin = 17
	BUTTONB Pin = 26
)

var DefaultUART = UART0

// UART pins
const (
	UART_TX_PIN Pin = 24
	UART_RX_PIN Pin = 25
)

// ADC pins
const (
	ADC0 Pin = 3 // P0 on the board
	ADC1 Pin = 2 // P1 on the board
	ADC2 Pin = 1 // P2 on the board
)

// I2C pins
const (
	SDA_PIN Pin = 30 // P20 on the board
	SCL_PIN Pin = 0  // P19 on the board
)

// SPI pins
const (
	SPI0_SCK_PIN Pin = 23 // P13 on the board
	SPI0_SDO_PIN Pin = 21 // P15 on the board
	SPI0_SDI_PIN Pin = 22 // P14 on the board
)

// GPIO/Analog pins
const (
	P0  Pin = 3
	P1  Pin = 2
	P2  Pin = 1
	P3  Pin = 4
	P4  Pin = 5
	P5  Pin = 17
	P6  Pin = 12
	P7  Pin = 11
	P8  Pin = 18
	P9  Pin = 10
	P10 Pin = 6
	P11 Pin = 26
	P12 Pin = 20
	P13 Pin = 23
	P14 Pin = 22
	P15 Pin = 21
	P16 Pin = 16
)

// LED matrix pins
const (
	LED_COL_1 Pin = 4
	LED_COL_2 Pin = 5
	LED_COL_3 Pin = 6
	LED_COL_4 Pin = 7
	LED_COL_5 Pin = 8
	LED_COL_6 Pin = 9
	LED_COL_7 Pin = 10
	LED_COL_8 Pin = 11
	LED_COL_9 Pin = 12
	LED_ROW_1 Pin = 13
	LED_ROW_2 Pin = 14
	LED_ROW_3 Pin = 15
)

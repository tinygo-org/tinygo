// +build nrf51,microbit

package machine

// The micro:bit does not have a 32kHz crystal on board.
const HasLowFrequencyCrystal = false

// Buttons on the micro:bit (A and B)
const (
	BUTTON  = BUTTONA
	BUTTONA = 17
	BUTTONB = 26
)

// UART pins
const (
	UART_TX_PIN = 24
	UART_RX_PIN = 25
)

// ADC pins
const (
	ADC0 = 3 // P0 on the board
	ADC1 = 2 // P1 on the board
	ADC2 = 1 // P2 on the board
)

// I2C pins
const (
	SDA_PIN = 30 // P20 on the board
	SCL_PIN = 0  // P19 on the board
)

// SPI pins
const (
	SPI0_SCK_PIN  = 23 // P13 on the board
	SPI0_MOSI_PIN = 21 // P15 on the board
	SPI0_MISO_PIN = 22 // P14 on the board
)

// GPIO/Analog pins
const (
	P0  = 3
	P1  = 2
	P2  = 1
	P3  = 4
	P4  = 5
	P5  = 17
	P6  = 12
	P7  = 11
	P8  = 18
	P9  = 10
	P10 = 6
	P11 = 26
	P12 = 20
	P13 = 23
	P14 = 22
	P15 = 21
	P16 = 16
)

// LED matrix pins
const (
	LED_COL_1 = 4
	LED_COL_2 = 5
	LED_COL_3 = 6
	LED_COL_4 = 7
	LED_COL_5 = 8
	LED_COL_6 = 9
	LED_COL_7 = 10
	LED_COL_8 = 11
	LED_COL_9 = 12
	LED_ROW_1 = 13
	LED_ROW_2 = 14
	LED_ROW_3 = 15
)

// +build nrf51,microbit

package machine

import (
	"device/nrf"
	"errors"
)

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
	ADC0 = 0
	ADC1 = 1
	ADC2 = 2
)

// I2C pins
const (
	SDA_PIN = 20
	SCL_PIN = 19
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

// matrixSettings has the legs of the LED grid in the form {row, column} for each LED position.
var matrixSettings = [5][5][2]uint8{
	{{LED_ROW_1, LED_COL_1}, {LED_ROW_2, LED_COL_4}, {LED_ROW_1, LED_COL_2}, {LED_ROW_2, LED_COL_5}, {LED_ROW_1, LED_COL_3}},
	{{LED_ROW_3, LED_COL_4}, {LED_ROW_3, LED_COL_5}, {LED_ROW_3, LED_COL_6}, {LED_ROW_3, LED_COL_7}, {LED_ROW_3, LED_COL_8}},
	{{LED_ROW_2, LED_COL_2}, {LED_ROW_1, LED_COL_9}, {LED_ROW_2, LED_COL_3}, {LED_ROW_3, LED_COL_9}, {LED_ROW_2, LED_COL_1}},
	{{LED_ROW_1, LED_COL_8}, {LED_ROW_1, LED_COL_7}, {LED_ROW_1, LED_COL_6}, {LED_ROW_1, LED_COL_5}, {LED_ROW_1, LED_COL_4}},
	{{LED_ROW_3, LED_COL_3}, {LED_ROW_2, LED_COL_7}, {LED_ROW_3, LED_COL_1}, {LED_ROW_2, LED_COL_6}, {LED_ROW_3, LED_COL_2}}}

// InitLEDMatrix initializes the LED matrix, by turning all of the row/col pins to output and high.
func InitLEDMatrix() {
	for i := LED_COL_1; i <= LED_ROW_3; i++ {
		nrf.GPIO.OUTSET = 1 << uint8(i)
		nrf.GPIO.DIRSET = 1 << uint8(i)
	}
}

// SetLEDMatrix sets a single LED on the LED matrix either on or off.
// Currently limited to a single LED at a time.
func SetLEDMatrix(x, y uint8, on bool) error {
	if x > 4 || y > 4 {
		return errors.New("Invalid LED matrix row or column")
	}

	nrf.GPIO.OUTCLR = (1 << 13)
	nrf.GPIO.OUTCLR = (1 << 14)
	nrf.GPIO.OUTCLR = (1 << 15)
	if on {
		nrf.GPIO.OUTSET = (1 << matrixSettings[y][x][0])
		nrf.GPIO.OUTCLR = (1 << matrixSettings[y][x][1])
	} else {
		nrf.GPIO.OUTCLR = (1 << matrixSettings[y][x][0])
		nrf.GPIO.OUTSET = (1 << matrixSettings[y][x][1])
	}

	return nil
}

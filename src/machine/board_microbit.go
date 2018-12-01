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
	SPI0_SCK_PIN  = 0
	SPI0_MOSI_PIN = 0
	SPI0_MISO_PIN = 0
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

// InitLEDMatrix initializes the LED matrix, by setting all of the row/col pins to output
// then calling ClearLEDMatrix.
func InitLEDMatrix() {
	set := 0
	for i := LED_COL_1; i <= LED_ROW_3; i++ {
		set |= 1 << uint8(i)
	}
	nrf.GPIO.DIRSET = nrf.RegValue(set)
	ClearLEDMatrix()
}

// ClearLEDMatrix clears the entire LED matrix.
func ClearLEDMatrix() {
	set := 0
	for i := LED_COL_1; i <= LED_COL_9; i++ {
		set |= 1 << uint8(i)
	}
	nrf.GPIO.OUTSET = nrf.RegValue(set)
	nrf.GPIO.OUTCLR = (1 << LED_ROW_1) | (1 << LED_ROW_2) | (1 << LED_ROW_3)
}

// SetLEDMatrix turns on a single LED on the LED matrix.
// Currently limited to a single LED at a time, it will clear the matrix before setting it.
func SetLEDMatrix(x, y uint8) error {
	if x > 4 || y > 4 {
		return errors.New("Invalid LED matrix row or column")
	}

	// Clear matrix
	ClearLEDMatrix()

	nrf.GPIO.OUTSET = (1 << matrixSettings[y][x][0])
	nrf.GPIO.OUTCLR = (1 << matrixSettings[y][x][1])

	return nil
}

// SetEntireLEDMatrixOn turns on all of the LEDs on the LED matrix.
func SetEntireLEDMatrixOn() error {
	set := 0
	for i := LED_ROW_1; i <= LED_ROW_3; i++ {
		set |= 1 << uint8(i)
	}
	nrf.GPIO.OUTSET = nrf.RegValue(set)

	set = 0
	for i := LED_COL_1; i <= LED_COL_9; i++ {
		set |= 1 << uint8(i)
	}
	nrf.GPIO.OUTCLR = nrf.RegValue(set)

	return nil
}

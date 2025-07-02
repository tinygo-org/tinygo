//go:build hw_651

package machine

// No-name brand board based on the nRF51822 chip with low frequency crystal on board.
// Pinout (reverse engineered from the board) can be found here:
// https://aviatorahmet.blogspot.com/2020/12/pinout-of-nrf51822-board.html
// https://cr0wg4n.medium.com/pinout-nrf51822-board-hw-651-78da2eda8894

const HasLowFrequencyCrystal = true

var DefaultUART = UART0

// GPIO pins on header J1
const (
	J1_01 = P0_21
	J1_03 = P0_23
	J1_04 = P0_22
	J1_05 = P0_25
	J1_06 = P0_24
	J1_09 = P0_29
	J1_10 = P0_28
	J1_11 = P0_30
	J1_13 = P0_00
	J1_15 = P0_02
	J1_17 = P0_04
	J1_16 = P0_01
	J1_18 = P0_03
)

// GPIO pins on header J2
const (
	J2_01 = P0_20
	J2_03 = P0_18
	J2_04 = P0_19
	J2_07 = P0_16
	J2_08 = P0_15
	J2_09 = P0_14
	J2_10 = P0_13
	J2_11 = P0_12
	J2_12 = P0_11
	J2_13 = P0_10
	J2_14 = P0_09
	J2_15 = P0_08
	J2_16 = P0_07
	J2_17 = P0_06
	J2_18 = P0_05
)

// UART pins
const (
	UART_TX_PIN = P0_24 // J1_06 on the board
	UART_RX_PIN = P0_25 // J1_05 on the board
)

// ADC pins
const (
	ADC0 = P0_03 // J1_18 on the board
	ADC1 = P0_02 // J1_15 on the board
	ADC2 = P0_01 // J1_16 on the board
	ADC3 = P0_04 // J1_17 on the board
	ADC4 = P0_05 // J2_18 on the board
	ADC5 = P0_06 // J2_17 on the board
)

// I2C pins
const (
	SDA_PIN = P0_30 // J1_11 on the board
	SCL_PIN = P0_00 // J1_13 on the board
)

// SPI pins
const (
	SPI0_SCK_PIN = P0_23 // J1_03 on the board
	SPI0_SDO_PIN = P0_21 // J1_01 on the board
	SPI0_SDI_PIN = P0_22 // J1_04 on the board
)

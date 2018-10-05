// +build nrf,microbit

package machine

// The micro:bit does not have a 32kHz crystal on board.
const HasLowFrequencyCrystal = false

// Buttons on the micro:bit (A and B)
const (
	BUTTON  = BUTTON1
	BUTTON1 = 5
	BUTTON2 = 11
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

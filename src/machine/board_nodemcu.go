//go:build nodemcu
// +build nodemcu

// Pinout for the NodeMCU dev kit.

package machine

// GPIO pins on the NodeMCU board.
const (
	D0 Pin = 16
	D1 Pin = 5
	D2 Pin = 4
	D3 Pin = 0
	D4 Pin = 2
	D5 Pin = 14
	D6 Pin = 12
	D7 Pin = 13
	D8 Pin = 15
)

// Onboard blue LED (on the AI-Thinker module).
const LED = D4

// SPI pins
const (
	SPI0_SCK_PIN = D5
	SPI0_SDO_PIN = D7
	SPI0_SDI_PIN = D6
	SPI0_CS0_PIN = D8
)

// I2C pins
const (
	SDA_PIN = D2
	SCL_PIN = D1
)

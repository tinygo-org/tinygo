//go:build nodemcu

// Pinout for the NodeMCU dev kit.

package machine

// GPIO pins on the NodeMCU board.
const (
	D0 = GPIO16
	D1 = GPIO5
	D2 = GPIO4
	D3 = GPIO0
	D4 = GPIO2
	D5 = GPIO14
	D6 = GPIO12
	D7 = GPIO13
	D8 = GPIO15
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

// +build feather_rp2040

package machine

const (
	LED = GPIO13

	// Onboard crystal oscillator frequency, in MHz.
	xoscFreq = 12 // MHz
)

// I2C Pins.
const (
	I2C0_SDA_PIN = GPIO24
	I2C0_SCL_PIN = GPIO25

	I2C1_SDA_PIN = GPIO2
	I2C1_SCL_PIN = GPIO3
)

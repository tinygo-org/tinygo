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

// SPI default pins
const (
	// Default Serial Clock Bus 0 for SPI communications
	SPI0_SCK_PIN = GPIO18
	// Default Serial Out Bus 0 for SPI communications
	SPI0_SDO_PIN = GPIO19 // Tx
	// Default Serial In Bus 0 for SPI communications
	SPI0_SDI_PIN = GPIO20 // Rx

	// Default Serial Clock Bus 1 for SPI communications
	SPI1_SCK_PIN = GPIO10
	// Default Serial Out Bus 1 for SPI communications
	SPI1_SDO_PIN = GPIO11 // Tx
	// Default Serial In Bus 1 for SPI communications
	SPI1_SDI_PIN = GPIO12 // Rx
)

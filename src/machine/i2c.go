// +build avr nrf

package machine

type I2C struct {
}

// TWI_FREQ is the bus speed. Normally either 100 kHz, or 400 kHz for high-speed bus.
const (
	TWI_FREQ_100KHZ = 100000
	TWI_FREQ_400KHZ = 400000
)

// I2CConfig does not do much of anything on Arduino.
type I2CConfig struct {
	Frequency uint32
}

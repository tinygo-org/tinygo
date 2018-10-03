// +build avr nrf

package machine

// TWI_FREQ is the I2C bus speed. Normally either 100 kHz, or 400 kHz for high-speed bus.
const (
	TWI_FREQ_100KHZ = 100000
	TWI_FREQ_400KHZ = 400000
)

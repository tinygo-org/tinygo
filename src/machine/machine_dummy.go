
// +build !avr,!nrf

package machine

// Dummy machine package, filled with no-ops.

type GPIOMode uint8

const (
	GPIO_INPUT  = iota
	GPIO_OUTPUT
)

const LED = 0

func (p GPIO) Configure(config GPIOConfig) {
}

func (p GPIO) Set(value bool) {
}

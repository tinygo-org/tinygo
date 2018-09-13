// +build !avr,!nrf

package machine

// Dummy machine package, filled with no-ops.

type GPIOMode uint8

const (
	GPIO_INPUT = iota
	GPIO_OUTPUT
)

// Fake LED numbers, for testing.
const (
	LED  = LED1
	LED1 = 0
	LED2 = 0
	LED3 = 0
	LED4 = 0
)

func (p GPIO) Configure(config GPIOConfig) {
}

func (p GPIO) Set(value bool) {
}

func (p GPIO) Get() (value bool) {
	return
}

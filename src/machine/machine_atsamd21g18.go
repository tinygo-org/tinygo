// +build atsamd21g18

package machine

// Peripheral abstraction layer for the atsamd21g18.

type GPIOMode uint8

const (
	portA = iota * 16
	portB
)

const (
	GPIO_INPUT  = 0
	GPIO_OUTPUT = 1
)

// Configure this pin with the given configuration.
func (p GPIO) Configure(config GPIOConfig) {
}

// Set the pin to high or low.
// Warning: only use this on an output pin!
func (p GPIO) Set(high bool) {
}

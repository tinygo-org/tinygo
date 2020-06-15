// +build k210

package machine

func CPUFrequency() uint32 {
	return 400000000
}

type PinMode uint8

const (
	PinInput PinMode = iota
	PinOutput
	PinPWM
	PinSPI
	PinI2C = PinSPI
)

// Configure this pin with the given configuration.
func (p Pin) Configure(config PinConfig) {
}

// Set the pin to high or low.
func (p Pin) Set(high bool) {
}

// Get returns the current value of a GPIO pin.
func (p Pin) Get() bool {
	return true
}

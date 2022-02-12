//go:build digispark
// +build digispark

package machine

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 16000000
}

// Digital pins
const (
	P0 Pin = PB0
	P1 Pin = PB1
	P2 Pin = PB2
	P3 Pin = PB3
	P4 Pin = PB4
	P5 Pin = PB5
)

// LED pin
const LED Pin = P1

// Analog pins
const (
	ADC0 Pin = PB5
	ADC1 Pin = PB2
	ADC2 Pin = PB4
	ADC3 Pin = PB3
)

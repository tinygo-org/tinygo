//go:build (rp2040 || (rp2350 && !rp2350b)) && !boardpackage

package machine

// Analog pins on RP2040 and RP2350A.
const (
	ADC0 Pin = GPIO26
	ADC1 Pin = GPIO27
	ADC2 Pin = GPIO28
	ADC3 Pin = GPIO29
)

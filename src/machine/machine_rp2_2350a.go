//go:build rp2350 && !rp2350b

package machine

// Analog pins on RP2350a.
const (
	ADC0 Pin = GPIO26
	ADC1 Pin = GPIO27
	ADC2 Pin = GPIO28
	ADC3 Pin = GPIO29

	// fifth ADC channel.
	thermADC = 30
)

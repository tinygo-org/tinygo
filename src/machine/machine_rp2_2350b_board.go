//go:build rp2350b && !boardpackage

package machine

// Analog pins on RP2350B.
const (
	ADC0 Pin = GPIO40
	ADC1 Pin = GPIO41
	ADC2 Pin = GPIO42
	ADC3 Pin = GPIO43
	ADC4 Pin = GPIO44
	ADC5 Pin = GPIO45
	ADC6 Pin = GPIO46
	ADC7 Pin = GPIO47
)

// Additional PWMs on the RP2350B.
var (
	PWM8  = getPWMGroup(8)
	PWM9  = getPWMGroup(9)
	PWM10 = getPWMGroup(10)
	PWM11 = getPWMGroup(11)
)

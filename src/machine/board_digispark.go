//go:build digispark

package machine

// Digispark is a tiny ATtiny85-based board with 6 I/O pins.
//
// PWM is available on the following pins:
//   - P0 (PB0): Timer0 channel A
//   - P1 (PB1): Timer0 channel B or Timer1 channel A (LED pin)
//   - P4 (PB4): Timer1 channel B
//
// Timer1 is recommended for PWM as it provides more flexible frequency control.

// Return the current CPU frequency in hertz.
func CPUFrequency() uint32 {
	return 16000000
}

const (
	P0 Pin = PB0 // PWM available (Timer0 OC0A)
	P1 Pin = PB1 // PWM available (Timer0 OC0B or Timer1 OC1A)
	P2 Pin = PB2
	P3 Pin = PB3
	P4 Pin = PB4 // PWM available (Timer1 OC1B)
	P5 Pin = PB5

	LED = P1
)

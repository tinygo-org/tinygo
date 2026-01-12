//go:build digispark

package main

import "machine"

var (
	// Use Timer1 for PWM (recommended for ATtiny85)
	pwm  = machine.Timer1
	pinA = machine.P1 // PB1, Timer1 channel A (LED pin)
	pinB = machine.P4 // PB4, Timer1 channel B
)

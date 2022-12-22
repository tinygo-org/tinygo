//go:build arduino

package main

import "machine"

var (
	// Configuration on an Arduino Uno.
	pwm  = machine.Timer2
	pinA = machine.PB3 // pin 11 on the Uno
	pinB = machine.PD3 // pin 3 on the Uno
)

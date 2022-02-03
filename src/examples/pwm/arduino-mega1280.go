//go:build arduino_mega1280
// +build arduino_mega1280

package main

import "machine"

var (
	// Configuration on an Arduino Uno.
	pwm  = machine.Timer3
	pinA = machine.PH3 // pin 6 on the Mega
	pinB = machine.PH4 // pin 7 on the Mega
)

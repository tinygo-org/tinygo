//go:build itsybitsy_m0

package main

import "machine"

var (
	pwm  = machine.TCC0
	pinA = machine.D3
	pinB = machine.D4
)

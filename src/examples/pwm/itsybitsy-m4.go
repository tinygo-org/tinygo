//go:build itsybitsy_m4

package main

import "machine"

var (
	pwm  = machine.TCC0
	pinA = machine.D12
	pinB = machine.D13
)

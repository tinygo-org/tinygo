//go:build pico
// +build pico

package main

import "machine"

var (
	pwm  = machine.PWM4 // Pin 25 (LED on pico) corresponds to PWM4.
	pinA = machine.LED
	pinB = machine.GPIO24
)

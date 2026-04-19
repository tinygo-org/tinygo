//go:build pico2 || pico_plus2

package main

import "machine"

var (
	pwm = machine.PWM0 // Pin 25 (LED on pico2) corresponds to PWM0.
	//pinA = machine.LED
	pinA = machine.GPIO16
)

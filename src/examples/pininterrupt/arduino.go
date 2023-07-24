//go:build arduino

package main

import "machine"

const (
	button          = machine.D2
	buttonMode      = machine.PinInputPullup
	buttonPinChange = machine.PinRising
)

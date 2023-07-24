//go:build circuitplay_express

package main

import "machine"

const (
	button          = machine.BUTTON
	buttonMode      = machine.PinInputPulldown
	buttonPinChange = machine.PinFalling
)

//go:build wioterminal

package main

import "machine"

const (
	button          = machine.BUTTON
	buttonMode      = machine.PinInput
	buttonPinChange = machine.PinFalling
)

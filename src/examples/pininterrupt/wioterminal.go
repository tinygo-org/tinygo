//go:build wioterminal

package main

import "machine"

const (
	buttonMode      = machine.PinInput
	buttonPinChange = machine.PinFalling
)

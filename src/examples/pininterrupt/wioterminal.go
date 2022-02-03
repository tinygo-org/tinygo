//go:build wioterminal
// +build wioterminal

package main

import "machine"

const (
	buttonMode      = machine.PinInput
	buttonPinChange = machine.PinFalling
)

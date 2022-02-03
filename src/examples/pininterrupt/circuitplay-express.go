//go:build circuitplay_express
// +build circuitplay_express

package main

import "machine"

const (
	buttonMode      = machine.PinInputPulldown
	buttonPinChange = machine.PinFalling
)

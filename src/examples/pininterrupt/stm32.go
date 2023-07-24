//go:build stm32

package main

import "machine"

const (
	button          = machine.BUTTON
	buttonMode      = machine.PinInputPulldown
	buttonPinChange = machine.PinRising | machine.PinFalling
)

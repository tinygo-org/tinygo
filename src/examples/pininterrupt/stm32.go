//go:build stm32

package main

import "machine"

const (
	buttonMode      = machine.PinInputPulldown
	buttonPinChange = machine.PinRising | machine.PinFalling
)

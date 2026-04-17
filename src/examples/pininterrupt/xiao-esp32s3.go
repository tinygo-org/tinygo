//go:build xiao_esp32s3

package main

import "machine"

const (
	button          = machine.D1
	buttonMode      = machine.PinInput
	buttonPinChange = machine.PinFalling
)

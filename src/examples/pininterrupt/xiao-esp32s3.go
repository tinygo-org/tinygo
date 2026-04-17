//go:build xiao_esp32s3

package main

import "machine"

const (
	button          = machine.D1
	buttonMode      = machine.PinInputPullup
	buttonPinChange = machine.PinFalling
)

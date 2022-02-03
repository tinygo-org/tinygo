//go:build pca10040
// +build pca10040

package main

import "machine"

const (
	buttonMode      = machine.PinInputPullup
	buttonPinChange = machine.PinRising
)

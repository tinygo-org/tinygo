//go:build rp2040

package main

import "machine"

const (
	TARGET_SCL = machine.GPIO25
	TARGET_SDA = machine.GPIO24
)

var (
	controller = machine.I2C1
	target     = machine.I2C0
)
